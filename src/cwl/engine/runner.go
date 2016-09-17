package cwl_engine

import (
	"cwl"
	"strings"
	"path/filepath"
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

func FileNameSplit(path string) (string,string) {
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, ".") {
		root, ext := FileNameSplit(filename[1:])
		return "." + root, ext
	}
	tmp := strings.Split(filename, ".")
	if len(tmp) == 1 {
		return tmp[0], ""
	}
	return strings.Join(tmp[:len(tmp)-1], "."), "." + tmp[len(tmp)-1]
}

func MapInputs(inputs cwl.JSONDict, mapper PathMapper) cwl.JSONDict {
	out := cwl.JSONDict{}
	for k, v := range inputs {
		if base, ok := v.(map[interface{}]interface{}); ok {
			if classBase, ok := base["class"]; ok {
				if classBase == "File" {
					x := cwl.JSONDict{"class": "File"}
					x["path"] = mapper.LocationToPath(base["location"].(string))
					root, ext := FileNameSplit(x["path"].(string))
					x["nameroot"] = root
					x["nameext"] = ext
					x["basename"] = filepath.Base(x["path"].(string))
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
