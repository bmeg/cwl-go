package cwl_engine

import (
	"cwl"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func NewLocalRunner(config Config) (JobRunner, error) {
	workdir, err := ioutil.TempDir(config.TmpdirPrefix, "cwlwork_")
	if err != nil {
		return nil, fmt.Errorf("Unable to create working dir")
	}
	return LocalRunner{Config: config, Workdir: workdir}, nil
}

type LocalRunner struct {
	Config  Config
	Workdir string
}

func (self LocalRunner) LocationToPath(location string) string {
	return location
}

func (self LocalRunner) GetWorkDirPath() string {
	return self.Workdir
}

func (self LocalRunner) Glob(pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(self.Workdir, pattern))
	return matches
}

func (self LocalRunner) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(self.Workdir, path))
}

func (self LocalRunner) StartProcess(inputs cwl.JSONDict, cmd_args []string, workdir, stdout, stderr, stdin, dockerImage string) (cwl.JSONDict, error) {

	cmd := exec.Command(cmd_args[0], cmd_args[1:]...)

	if stdout != "" {
		var err error
		cmd.Stdout, err = os.Create(filepath.Join(workdir, stdout))
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	if stderr != "" {
		var err error
		cmd.Stderr, err = os.Create(filepath.Join(workdir, stderr))
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	if stdin != "" {
		var err error
		cmd.Stdin, err = os.Open(stdin)
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	cmd.Dir = workdir

	resFile := self.Workdir + ".result"

	log.Printf("Workdir: %s", workdir)
	go func(resfile string) {
		cmd_err := cmd.Run()
		exitStatus := 0
		if exiterr, ok := cmd_err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
				log.Printf("Exit Status: %d", exitStatus)
			}
		} else {
			log.Printf("cmd.Run: %v", cmd_err)
		}
		ioutil.WriteFile(resfile, []byte(fmt.Sprintf("%d", exitStatus)), 0600)
	}(resFile)
	return cwl.JSONDict{"resFile": resFile}, nil
}

func (self LocalRunner) GetOutput(prodData cwl.JSONDict) cwl.JSONDict {
	out := cwl.JSONDict{}
	path := filepath.Join(self.Workdir, "cwl.output.json")
	if _, err := os.Stat(path); err == nil {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(path)
		err := yaml.Unmarshal(data, &out)
		log.Printf("Returned: %s = %s %s", data, out, err)
	}
	return out
}

func (self LocalRunner) ExitCode(procData cwl.JSONDict) (int, bool) {
	if _, err := os.Stat(procData["resFile"].(string)); err == nil {
		d, _ := ioutil.ReadFile(procData["resFile"].(string))
		i, _ := strconv.Atoi(string(d))
		return i, true
	}
	return 0, false
}
