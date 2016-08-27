
package cwl_engine

import (
  "os"
  "cwl"
  "os/exec"
)


func NewLocalRunner(config cwl.Config) CWLRunner {
  return LocalRunner{}
}

type LocalRunner struct {
  
}

func (self LocalRunner) RunCommand(job cwl.Job) (cwl.JSONDict, error) {
    cmd := exec.Command(job.Cmd[0], job.Cmd[1:]...)
    cmd.Stderr = os.Stderr
    cmd.Stdout = os.Stdout
    err := cmd.Run()
    return cwl.JSONDict{}, err
}