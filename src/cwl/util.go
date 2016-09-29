package cwl

import (
	"encoding/json"
	"log"
	"strings"
)

func (self JSONDict) ToString() []byte {
	a := self.Normalize()
	o, err := json.Marshal(a)
	if err != nil {
		log.Printf("Output Error: %s", err)
	}
	return o
}

func mapNormalize(v interface{}) interface{} {
	if base, ok := v.(map[interface{}]interface{}); ok {
		out := map[string]interface{}{}
		for k, v := range base {
			out[k.(string)] = mapNormalize(v)
		}
		return out
	} else if base, ok := v.(JSONDict); ok {
		out := map[string]interface{}{}
		for k, v := range base {
			out[k.(string)] = mapNormalize(v)
		}
		return out

	} else if base, ok := v.([]interface{}); ok {
		out := make([]interface{}, len(base))
		for i, v := range base {
			out[i] = mapNormalize(v)
		}
		return out
	}
	return v
}

func (self JSONDict) Normalize() map[string]interface{} {
	return mapNormalize(self).(map[string]interface{})
}

func getFilePaths(x interface{}) []string {
	out := []string{}
	if a, ok := x.(JSONDict); ok {
		if c, ok := a["class"]; ok {
			if c.(string) == "File" {
				out = append(out, a["path"].(string))
			}
		} else {
			for _, v := range a {
				out = append(out, getFilePaths(v)...)
			}
		}
	}
	if a, ok := x.(map[interface{}]interface{}); ok {
		if c, ok := a["class"]; ok {
			if c.(string) == "File" {
				out = append(out, a["path"].(string))
			}
		} else {
			for _, v := range a {
				out = append(out, getFilePaths(v)...)
			}
		}
	}
	if a, ok := x.([]interface{}); ok {
		for _, i := range a {
			out = append(out, getFilePaths(i)...)
		}
	}
	return out
}

func (self JSONDict) GetFilePaths() []string {
	return getFilePaths(self)
}

func (self GraphState) HasResults(stepId string) bool {
	if _, ok := self[stepId]; ok {
		if _, ok := self[stepId][RESULTS_FIELD]; ok {
			return true
		}
	}
	return false
}

func (self GraphState) GetData(path string) (interface{}, bool) {
	tmp := strings.Split(path, "/")
	if len(tmp) == 1 {
		if base, ok := self[INPUT_FIELD]; ok {
			if strings.HasPrefix(path, "#") {
				path = path[1:]
			}
			if v, ok := (base[RESULTS_FIELD].(JSONDict))[path]; ok {
				return v, true
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	} else {
		if strings.HasPrefix(tmp[0], "#") {
			tmp[0] = tmp[0][1:]
		}
		if base, ok := self[tmp[0]]; ok {
			if _, ok := base[RESULTS_FIELD]; ok {
				if v, ok := (base[RESULTS_FIELD].(JSONDict))[tmp[1]]; ok {
					return v, true
				} else {
					return nil, false
				}
			} else {
				return nil, false
			}
		}
	}
	return nil, false
}

func (self *JobFile) ToJSONDict() JSONDict {
	a := JSONDict{
		"class":    "File",
		"location": self.Location,
	}
	if self.LoadContents {
		a["loadContents"] = true
	}
	return a
}
