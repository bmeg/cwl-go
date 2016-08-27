
package cwl_engine


import (
  "cwl"
)

type CWLRunner interface {
  RunCommand(cwl.Job) (cwl.JSONDict, error)
}
