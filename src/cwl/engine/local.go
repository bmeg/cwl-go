
package cwl_engine

import (
  "cwl"
)

type CWLRunner interface {
  RunCommand(cwl.CWLDoc, cwl.JSONDict)
}


func NewLocalRunner(config cwl.Config) CWLRunner {
  return LocalRunner{}
}

type LocalRunner struct {
  
}

func (self LocalRunner) RunCommand(doc cwl.CWLDoc, inputs cwl.JSONDict) {
  
}