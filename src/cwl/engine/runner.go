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

type PathMapper interface {
	LocationToPath(location string) string
}

func MapInputs(inputs cwl.JSONDict, mapper PathMapper) cwl.JSONDict {
	out := cwl.JSONDict{}
	for k, v := range inputs {
		if base, ok := v.(map[interface{}]interface{}); ok {
			if classBase, ok := base["class"]; ok {
				if classBase == "File" {
					x := cwl.JSONDict{"class": "File"}
					x["path"] = mapper.LocationToPath(base["location"].(string))
					out[k] = x
				}
				if classBase == "Directory" {
					x := cwl.JSONDict{"class": "Directory"}
					x["path"] = mapper.LocationToPath(base["location"].(string))
					out[k] = x
				}
			}
		}
	}
	return out
}
