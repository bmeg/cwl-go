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

func (self GraphState) HasResults(stepId string) bool {
	if _, ok := self[stepId]; ok {
		if _, ok := self[stepId][RESULTS_FIELD]; ok {
			return true
		}
	}
	return false
}

func (self GraphState) HasData(path string) bool {
	tmp := strings.Split("/", path)
	if len(tmp) == 1 {
		if base, ok := self[INPUT_FIELD]; ok {
			if _, ok := (base[RESULTS_FIELD].(JSONDict))[path]; ok {
				return true
			}
		} else {
			log.Printf("Weird GraphState data, %#v", self["#"])
		}
	} else {
		if base, ok := self[tmp[0]]; ok {
			if _, ok := base[RESULTS_FIELD]; ok {
				if _, ok := (base[RESULTS_FIELD].(JSONDict))[tmp[1]]; ok {
					return true
				}
			}
		}
	}
	return false
}

func (self GraphState) GetData(path string) interface{} {
	tmp := strings.Split("/", path)
	if len(tmp) == 1 {
		if base, ok := self[INPUT_FIELD]; ok {
			if v, ok := (base[RESULTS_FIELD].(JSONDict))[path]; ok {
				return v
			}
		} else {
			log.Printf("Weird GraphState data, %#v", self["#"])
		}
	} else {
		if base, ok := self[tmp[0]]; ok {
			if _, ok := base[RESULTS_FIELD]; ok {
				if v, ok := (base[RESULTS_FIELD].(JSONDict))[tmp[1]]; ok {
					return v
				}
			}
		}
	}
	return nil
}

func (self *JobFile) ToJSONDict() JSONDict {
	return JSONDict{
		"class":    "File",
		"location": self.Location,
	}
}
