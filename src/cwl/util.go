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

func IsFileStruct(x interface{}) bool {
	if base, ok := x.(map[interface{}]interface{}); ok {
		if b, ok := base["class"]; ok {
			if b == "File" {
				return true
			}
		}
	}
	return false
}

func (self JSONDict) GetFilePaths() []string {
	return getFilePaths(self)
}

func (self JSONDict) GetData(path string) (interface{}, bool) {
	//log.Printf("Get Search: %s", path)
	tmp := strings.Split(path, "/")
	if len(tmp) == 1 {
		/*
			if strings.HasPrefix(path, "#") {
				path = path[1:]
			}
		*/
		b, ok := self[path]
		return b, ok
	} else {
		/*
			if strings.HasPrefix(tmp[0], "#") {
				tmp[0] = tmp[0][1:]
			}
		*/
		if base, ok := self[tmp[0]]; ok {
			if b, ok := base.(JSONDict); ok {
				return b.GetData(strings.Join(tmp[1:], "/"))
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
