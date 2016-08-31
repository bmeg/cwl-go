package cwl_engine

import (
	"cwl"
	//"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	log.Printf("Command Files: %#v", job.Files)
	log.Printf("Command Inputs: %#v", job.Inputs)

	inputs := MapInputs(job.Inputs, self)

	workdir, _ := ioutil.TempDir(self.Config.TmpdirPrefix, "cwlwork_")
	log.Printf("Command Args: %s", job.Cmd)

	cmd_args := make([]string, len(job.Cmd))
	for i := range job.Cmd {
		cmd_args[i], _ = cwl.ExpressionEvaluate(job.Cmd[i], inputs)
	}

	cmd := exec.Command(cmd_args[0], cmd_args[1:]...)

	if job.Stdout != "" {
		stdout, _ := cwl.ExpressionEvaluate(job.Stdout, inputs)
		cmd.Stdout, _ = os.Create(filepath.Join(workdir, stdout))
	}
	if job.Stderr != "" {
		stderr, _ := cwl.ExpressionEvaluate(job.Stderr, inputs)
		cmd.Stderr, _ = os.Create(filepath.Join(workdir, stderr))
	}
	if job.Stdin != "" {
		stdin, _ := cwl.ExpressionEvaluate(job.Stdin, inputs)
		log.Printf("STDIN: %s", stdin)
		var err error
		cmd.Stdin, err = os.Open(stdin)
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	cmd.Dir = workdir
	log.Printf("Workdir: %s", workdir)
	err := cmd.Run()

	if _, err := os.Stat(filepath.Join(workdir, "cwl.output.json")); !os.IsNotExist(err) {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(filepath.Join(workdir, "cwl.output.json"))
		out := map[interface{}]interface{}{}
		//err := json.Unmarshal(data, &out)
		err := yaml.Unmarshal(data, &out)
		log.Printf("Returned: %s = %s %s", data, out, err)
		return out, err
	}

	out := cwl.JSONDict{}
	for _, o := range job.Files {
		if o.Output {
			if o.Glob != "" {
				log.Printf("Output %s", o.Glob)
				g, _ := filepath.Glob(filepath.Join(workdir, o.Glob))
				for _, p := range g {
					log.Printf("Found %s", p)
				}
				f := cwl.JSONDict{"path": g[0]}
				out[o.Id] = f
			}
		}
	}
	return out, err
}
