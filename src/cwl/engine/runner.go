package cwl_engine

import (
	"cwl"
)

type Config struct {
	TmpOutdirPrefix string
	TmpdirPrefix    string
	Outdir          string
	Quiet           bool
}

type CWLRunner interface {
	RunCommand(cwl.Job) (cwl.JSONDict, error)
}
