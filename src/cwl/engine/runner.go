package cwl_engine

import (
	"crypto/sha1"
	"cwl"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	TmpOutdirPrefix string
	TmpdirPrefix    string
	Outdir          string
	Quiet           bool
}

type TaskRecord struct {
	ProcData cwl.JSONDict
	Inputs   cwl.JSONDict
	Stderr   string
	Stdout   string
	Workdir  string
	Job      cwl.Job
}

type PathMapper interface {
	MapFile(map[interface{}]interface{}) map[interface{}]interface{}
}

type JobRunner interface {
	StartProcess(inputs cwl.JSONDict, cmd_args []string, workdir, stdout, stderr, stdin, dockerImage string) (cwl.JSONDict, error)
	ExitCode(prodData cwl.JSONDict) (int, bool)
	GetOutput(prodData cwl.JSONDict) cwl.JSONDict
	GetWorkDirPath() string
	Glob(path string) []string
	ReadFile(path string) ([]byte, error)
}

func FileNameSplit(path string) (string, string) {
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, ".") && len(filename) > 1 {
		root, ext := FileNameSplit(filename[1:])
		return "." + root, ext
	}
	tmp := strings.Split(filename, ".")
	if len(tmp) == 1 {
		return tmp[0], ""
	}
	return strings.Join(tmp[:len(tmp)-1], "."), "." + tmp[len(tmp)-1]
}

type RuntimeMapper struct {
	Runner JobRunner
}


func (self RuntimeMapper) MapFile(in map[interface{}]interface{}) map[interface{}]interface{} {
	out := in
	if in["class"].(string) == "File" {
		root, ext := FileNameSplit(in["path"].(string))
		out["nameroot"] = root
		out["nameext"] = ext
		out["basename"] = filepath.Base(in["path"].(string))
		if b, ok := in["loadContents"]; ok {
			if b.(bool) {
				out["contents"], _ = self.Runner.ReadFile(in["location"].(string))
				log.Printf("Load Contents: %s", in["location"].(string))
			}
		}
	}
	return out
}


func mapInputs(x interface{}, mapper PathMapper) interface{} {
	if base, ok := x.(map[interface{}]interface{}); ok {
		if classBase, ok := base["class"]; ok {
			if (classBase == "File" || classBase == "Directory") && len(base["location"].(string)) > 0 {
				return mapper.MapFile(base)
			}
		}
		out := map[interface{}]interface{}{}
		for k, v := range base {
			out[k] = mapInputs(v, mapper)
		}
	}

	if base, ok := x.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, mapInputs(i, mapper))
		}
		return out
	}
	return x
}

func MapInputs(inputs cwl.JSONDict, mapper PathMapper) cwl.JSONDict {
	out := cwl.JSONDict{}
	for k, v := range inputs {
		out[k] = mapInputs(v, mapper)
	}
	return out
}

func StartJob(job cwl.Job, runner JobRunner, pathMapper PathMapper) (TaskRecord, error) {

	log.Printf("Command Args: %#v", job.Cmd)
	log.Printf("Command Files: %#v", job.GetFiles())
	log.Printf("Command Inputs: %#v", job.InputData)
	log.Printf("Command Outputs: %#v", job.Outputs)

	input_data := job.InputData
	//attempting to get input files not mentioned in the user request, ie
	//default files. Not sure if this covers all cases
	for _, i := range job.GetFiles() {
		if !i.Output {
			if i.Id != "" {
				//if _, ok := input_data[i.Id]; !ok {
				log.Printf("Translating %s %#v", i.Id, i)
				input_data[i.Id] = map[interface{}]interface{}{
					"class":        "File",
					"location":     i.Location,
					"path":         i.Path,
					"loadContents": i.LoadContents,
				}
				//}
			}
		}
	}
	log.Printf("Translated Input: %s", input_data)
	runtimeMapper := RuntimeMapper{Runner:runner}
	//get the inputs using the path mapper from the job runner
	inputs := MapInputs(input_data, runtimeMapper)
	log.Printf("Mapped Inputs: %s", inputs)
	js_eval := cwl.JSEvaluator{Inputs: inputs}
	//process command line arguments
	cmd_args := []string{}
	if job.JobType == cwl.COMMAND {
		for i := range job.Cmd {
			s, err := job.Cmd[i].GetArgs(js_eval, func(x interface{}) interface{} {
				return mapInputs(x, pathMapper)
			})
			if err != nil {
				return TaskRecord{}, fmt.Errorf("Expression Eval failed: %s", err)
			}
			cmd_args = append(cmd_args, s...)
		}
	} else if job.JobType == cwl.EXPRESSION {
		cmd_args = append(cmd_args, job.Expression)
		js_inputs := cwl.JSONDict{}
		for i := range job.Cmd {
			s, err := job.Cmd[i].EvaluateObject(js_eval)
			if err != nil {
				return TaskRecord{}, fmt.Errorf("Input Eval Failure: %s", err)
			}
			js_inputs[job.Cmd[i].Id] = s
		}
		inputs = MapInputs(js_inputs, pathMapper)
		log.Printf("Expression Conversion: %s", inputs)
	}
	log.Printf("CMD: %s", cmd_args)

	stdout := ""
	stderr := ""
	stdin := ""
	workdir := runner.GetWorkDirPath()
	if job.Stdout != "" {
		var err error
		stdout, err = js_eval.EvaluateExpressionString(job.Stdout, nil)
		if err != nil {
			return TaskRecord{}, err
		}
	}
	if job.Stderr != "" {
		var err error
		stderr, err = js_eval.EvaluateExpressionString(job.Stderr, nil)
		if err != nil {
			return TaskRecord{}, err
		}
	}
	if job.Stdin != "" {
		var err error
		stdin, err = js_eval.EvaluateExpressionString(job.Stdin, nil)
		if err != nil {
			return TaskRecord{}, err
		}
	}

	proc_data, err := runner.StartProcess(inputs, cmd_args, workdir, stdout, stderr, stdin, job.DockerImage)
	return TaskRecord{ProcData: proc_data, Workdir: workdir, Inputs: inputs, Stdout: stdout, Stderr: stderr, Job: job}, err
}

func CleanupJob(task_data TaskRecord, runner JobRunner) (cwl.JSONDict, error) {
	out := runner.GetOutput(task_data.ProcData)

	js_eval := cwl.JSEvaluator{Inputs: task_data.Inputs}

	out_files := cwl.JSONDict{}
	for _, o := range task_data.Job.GetFiles() {
		if o.Output {
			if o.Glob != "" {
				glob, _ := js_eval.EvaluateExpressionString(o.Glob, nil)
				log.Printf("Output File Glob: %s", glob)
				g := runner.Glob(glob)
				for _, p := range g {
					log.Printf("Found %s %s", o.Id, p)
					hasher := sha1.New()
					file, _ := os.Open(p)
					if _, err := io.Copy(hasher, file); err != nil {
						log.Fatal(err)
					}
					hash_val := fmt.Sprintf("sha1$%x", hasher.Sum([]byte{}))
					file.Close()
					info, _ := os.Stat(p)
					f := map[interface{}]interface{}{"location": p, "checksum": hash_val, "class": "File", "size": info.Size()}
					out_files[o.Id] = f
				}
			}
		}
	}
	for k, v := range task_data.Job.Outputs {
		if v.TypeName == "File" || v.TypeName == "stdout" || v.TypeName == "stderr" {
			if _, ok := out_files[k]; ok {
				out[k] = out_files[k]
			}
		}
	}

	return out, nil
}

func JobDone(task_data TaskRecord, runner JobRunner) bool {
	_, done := runner.ExitCode(task_data.ProcData)
	return done
}
