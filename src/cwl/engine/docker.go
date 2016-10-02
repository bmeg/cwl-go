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
	"strings"
	"syscall"
)

func NewDockerRunner(config Config) (JobRunner, error) {
	workdir, err := ioutil.TempDir(config.TmpdirPrefix, "cwlwork_")
	if err != nil {
		return nil, fmt.Errorf("Unable to create working dir")
	}
	return DockerRunner{Config: config, hostWorkDir: workdir, fileMap: map[string]string{}, fileMapRev: map[string]string{}}, nil
}

type DockerRunner struct {
	Config      Config
	hostWorkDir string
	fileMap     map[string]string
	fileMapRev  map[string]string
}

func (self DockerRunner) GetWorkDirPath() string {
	return "/var/run/cwl-go"
}

func (self DockerRunner) LocationToPath(location string) string {
	if out, ok := self.fileMapRev[location]; ok {
		return out
	}
	base := filepath.Base(location)
	out := fmt.Sprintf("/var/run/cwlinput/%d/%s", len(self.fileMap), base)
	self.fileMap[out] = location
	self.fileMapRev[location] = out
	log.Printf("Translating: %s to %s", location, out)
	return out
}

func (self DockerRunner) StartProcess(inputs cwl.JSONDict, cmd_args []string, workdir, stdout, stderr, stdin, dockerImage string) (cwl.JSONDict, error) {

	binds := []string{fmt.Sprintf("%s:%s", self.hostWorkDir, workdir)}

	log.Printf("Docker files: %s", inputs.GetFilePaths())

	for _, n := range inputs.GetFilePaths() {
		binds = append(binds, fmt.Sprintf("%s:%s", self.fileMap[n], n))
	}
	log.Printf("Docker Binds: %s", binds)

	resFile := self.hostWorkDir + ".result"
	go func(callback func(int)) {

		args := []string{"run", "--rm", "-i", "-w", workdir}

		for _, i := range binds {
			args = append(args, "-v", i)
		}
		args = append(args, dockerImage)
		args = append(args, cmd_args...)
		log.Printf("Runner docker %s", strings.Join(args, " "))

		cmd := exec.Command("docker", args...)

		if stdout != "" {
			var err error
			cmd.Stdout, err = os.Create(filepath.Join(self.hostWorkDir, stdout))
			if err != nil {
				callback(1)
				return
			}
		}
		if stderr != "" {
			var err error
			cmd.Stderr, err = os.Create(filepath.Join(self.hostWorkDir, stderr))
			if err != nil {
				callback(1)
				return
			}
		}
		if stdin != "" {
			log.Printf("Stdin %s to %s", stdin, self.fileMap[stdin])
			var err error
			cmd.Stdin, err = os.Open(self.fileMap[stdin])
			if err != nil {
				callback(1)
				return
			}
		}
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
		callback(0)
	}(func(exitStatus int) {
		ioutil.WriteFile(resFile, []byte(fmt.Sprintf("%d", exitStatus)), 0600)
	})
	return cwl.JSONDict{"resFile": resFile}, nil
}

func (self DockerRunner) GetOutput(prodData cwl.JSONDict) cwl.JSONDict {
	out := cwl.JSONDict{}
	path := filepath.Join(self.hostWorkDir, "cwl.output.json")
	if _, err := os.Stat(path); err == nil {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(path)
		err := yaml.Unmarshal(data, &out)
		log.Printf("Returned: %s = %s %s", data, out, err)
	}
	return out
}

func (self DockerRunner) ReadFile(location string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(location))
}

func (self DockerRunner) Glob(pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(self.hostWorkDir, pattern))
	return matches
}

func (self DockerRunner) ExitCode(procData cwl.JSONDict) (int, bool) {
	if _, err := os.Stat(procData["resFile"].(string)); err == nil {
		d, _ := ioutil.ReadFile(procData["resFile"].(string))
		i, _ := strconv.Atoi(string(d))
		return i, true
	}
	return 0, false
}
