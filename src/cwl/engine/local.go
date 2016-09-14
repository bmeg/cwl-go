package cwl_engine

import (
	"crypto/sha1"
	"cwl"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func NewLocalRunner(config Config) CWLRunner {
	return LocalRunner{Config: config}
}

type LocalRunner struct {
	Config Config
}

func (self LocalRunner) LocationToPath(location string) string {
	return location
}

func (self LocalRunner) RunCommand(job cwl.Job) (cwl.JSONDict, error) {
	log.Printf("Command Files: %#v", job.GetFiles)
	log.Printf("Command Inputs: %#v", job.InputData)

	inputs := MapInputs(job.InputData, self)

	workdir, err := ioutil.TempDir(self.Config.TmpdirPrefix, "cwlwork_")
	if err != nil {
		return cwl.JSONDict{}, fmt.Errorf("Unable to create working dir")
	}
	log.Printf("Command Args: %#v", job.Cmd)

	cmd_args := []string{}

	js_eval := cwl.JSEvaluator{Inputs: inputs}

	for i := range job.Cmd {
		s, err := job.Cmd[i].EvaluateStrings(js_eval)
		if err != nil {
			return cwl.JSONDict{}, fmt.Errorf("Expression Eval failed: %s", err)
		}
		cmd_args = append(cmd_args, s...)
	}
	log.Printf("CMD: %s", cmd_args)
	cmd := exec.Command(cmd_args[0], cmd_args[1:]...)

	if job.Stdout != "" {
		stdout, _ := js_eval.EvaluateExpressionString(job.Stdout, nil)
		cmd.Stdout, _ = os.Create(filepath.Join(workdir, stdout))
	}
	if job.Stderr != "" {
		stderr, _ := js_eval.EvaluateExpressionString(job.Stderr, nil)
		cmd.Stderr, _ = os.Create(filepath.Join(workdir, stderr))
	}
	if job.Stdin != "" {
		stdin, _ := js_eval.EvaluateExpressionString(job.Stdin, nil)
		log.Printf("STDIN: %s", stdin)
		var err error
		cmd.Stdin, err = os.Open(stdin)
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	cmd.Dir = workdir
	log.Printf("Workdir: %s", workdir)
	cmd_err := cmd.Run()
	if exiterr, ok := cmd_err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			exitStatus := status.ExitStatus()
			log.Printf("Exit Status: %d", exitStatus)
			found := false
			for _, i := range job.SuccessCodes {
				if i == exitStatus {
					found = true
				}
			}
			if !found {
				return cwl.JSONDict{}, cmd_err
			}
			cmd_err = nil
		}
	} else {
		log.Printf("cmd.Run: %v", err)
	}

	meta := map[interface{}]interface{}{}
	if _, err := os.Stat(filepath.Join(workdir, "cwl.output.json")); !os.IsNotExist(err) {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(filepath.Join(workdir, "cwl.output.json"))
		//err := json.Unmarshal(data, &out)
		err := yaml.Unmarshal(data, &meta)
		log.Printf("Returned: %s = %s %s", data, meta, err)
		return meta, nil
	}

	out := cwl.JSONDict{}
	for _, o := range job.GetFiles() {
		if o.Output {
			if o.Glob != "" { //BUG: should actually type check the schema declaration here
				log.Printf("Output %s", filepath.Join(workdir, o.Glob))
				g, _ := filepath.Glob(filepath.Join(workdir, o.Glob))
				for _, p := range g {
					log.Printf("Found %s", p)
					hasher := sha1.New()
					file, _ := os.Open(p)
					if _, err := io.Copy(hasher, file); err != nil {
						log.Fatal(err)
					}
					hash_val := fmt.Sprintf("sha1$%x", hasher.Sum([]byte{}))
					file.Close()
					info, _ := os.Stat(p)
					f := map[interface{}]interface{}{"location": p, "checksum": hash_val, "class": "File", "size": info.Size()}
					out[o.Id] = f
				}
			} else {
				if a, ok := meta[o.Id]; ok {
					out[o.Id] = a
				}
			}
		}
	}
	return out, cmd_err
}
