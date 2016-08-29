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

func (self LocalRunner) RunCommand(job cwl.Job) (cwl.JSONDict, error) {
	log.Printf("Files: %#v", job.Files)
	workdir, _ := ioutil.TempDir(self.Config.TmpdirPrefix, "cwlwork_")
	cmd := exec.Command(job.Cmd[0], job.Cmd[1:]...)

	if job.Stdout != "" {
		cmd.Stdout, _ = os.Create(filepath.Join(workdir, job.Stdout))
	}
	if job.Stderr != "" {
		cmd.Stderr, _ = os.Create(filepath.Join(workdir, job.Stderr))
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

	for _, o := range job.Files {
		if o.Output {
			if o.Glob != "" {
				log.Printf("Output %s", o.Glob)
				g, _ := filepath.Glob(filepath.Join(workdir, o.Glob))
				for _, p := range g {
					log.Printf("Found %s", p)
				}
			}
		}
	}

	return cwl.JSONDict{}, err
}
