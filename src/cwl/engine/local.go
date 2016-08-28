package cwl_engine

import (
	"cwl"
	"encoding/json"
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
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = workdir
	log.Printf("Workdir: %s", workdir)
	err := cmd.Run()

	if _, err := os.Stat(filepath.Join(workdir, "cwl.output.json")); !os.IsNotExist(err) {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(filepath.Join(workdir, "cwl.output.json"))
		out := cwl.JSONDict{}
		err := json.Unmarshal(data, &out)
		log.Printf("Returned: %s = %s %s", data, out, err)
		return out, nil
	}
	return cwl.JSONDict{}, err
}
