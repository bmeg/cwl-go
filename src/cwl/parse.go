package cwl

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
)

func InputParse(path string) (JSONDict, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := make(JSONDict)
	err = yaml.Unmarshal(source, &doc)
	return doc, err
}

func (self *JSONDict) Write(o io.Writer) {
	jout, err := json.Marshal(self)
	if err != nil {
		return
	}
	o.Write(jout)
}

func Parse(cwl_path string) (CWLDoc, error) {
	source, err := ioutil.ReadFile(cwl_path)
	if err != nil {
		panic(err)
	}
	doc := make(CWLDocData)
	err = yaml.Unmarshal(source, &doc)
	x, _ := filepath.Abs(cwl_path)
	base_path := filepath.Dir(x)
	if doc["class"].(string) == "CommandLineTool" {
		return NewCommandLineTool(doc, base_path)
	}
	return nil, nil
}

func NewCommandLineTool(doc CWLDocData, cwl_path string) (CWLDoc, error) {
	log.Printf("CommandLineTool: %v", doc)
	out := CommandLineTool{}
	out.Inputs = make(map[string]CommandInput)
	out.Outputs = make(map[string]CommandOutput)

	if _, ok := doc["id"]; ok {
		out.Id = doc["id"].(string)
	} else {
		out.Id = ""
	}

	if base, ok := doc["requirements"]; ok {
		r, err := NewRequirements(base)
		if err != nil {
			return CommandLineTool{}, err
		}
		log.Printf("Requirements: %#v", r)
		out.Requirements = r
	}

	if base, ok := doc["baseCommand"].([]string); ok {
		out.BaseCommand = base
	} else {
		if base, ok := doc["baseCommand"].(string); ok {
			out.BaseCommand = []string{base}
		}
	}

	if base, ok := doc["arguments"]; ok {
		for _, x := range base.([]interface{}) {
			n, err := NewArgument(x)
			if err == nil {
				out.Arguments = append(out.Arguments, n)
			}
		}
	}

	if base, ok := doc["inputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := NewCommandInput(k.(string), v, cwl_path)
				if err == nil {
					out.Inputs[n.Id] = n
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Input array: %d", len(base_array))
			for _, x := range base_array {
				n, err := NewCommandInput("", x, cwl_path)
				if err == nil {
					out.Inputs[n.Id] = n
				} else {
					log.Printf("Command line Input error: %s", err)
				}
			}
		} else {
			log.Printf("Can't Parse Inputs")
		}
	} else {
		log.Printf("No Inputs found")
	}

	log.Printf("Parse CommandLineTool: %v", out)
	return out, nil
}

func NewCommandInput(id string, x interface{}, cwl_path string) (CommandInput, error) {
	out := CommandInput{}
	if base, ok := x.(map[interface{}]interface{}); ok {
		if id == "" {
			out.Id = base["id"].(string)
		} else {
			out.Id = id
		}
		if binding, ok := base["inputBinding"]; ok {
			if pos, ok := binding.(map[interface{}]interface{})["position"]; ok {
				out.Position = pos.(int)
			} else {
				out.Position = 100000
			}
			if prefix, ok := binding.(map[interface{}]interface{})["prefix"].(string); ok {
				out.Prefix = &prefix
			}
			if itemSep, ok := binding.(map[interface{}]interface{})["itemSeparator"].(string); ok {
				out.ItemSeparator = &itemSep
			}
		}
		t, err := NewDataType(base["type"])
		if err != nil {
			return out, fmt.Errorf("unable to load data type: %s", err)
		}
		out.Type = t
		if def, ok := base["default"]; ok {
			out.Default = &def
			//special case when default value is a file
			if out.Type.TypeName == "File" {
				if base, ok := def.(map[interface{}]interface{}); ok {
					if s, ok := base["path"]; ok {
						base["path"] = path.Join(cwl_path, s.(string))
					}
					if s, ok := base["location"]; ok {
						base["location"] = path.Join(cwl_path, s.(string))
					}
				}
			}
		}

	} else {
		return out, fmt.Errorf("Unable to parse CommandInput: %v", x)
	}
	log.Printf("CommandInput: %#v", out)
	return out, nil
}

func NewDataType(value interface{}) (DataType, error) {
	if base, ok := value.(string); ok {
		return DataType{TypeName: base}, nil
	} else if base, ok := value.(map[interface{}]interface{}); ok {
		out := DataType{TypeName: base["type"].(string)}
		//in the case of an item array, there can be command line bindings for the elements
		if bBase, bOk := base["inputBinding"]; bOk {
			if binding, bOk := bBase.(map[interface{}]interface{}); bOk {
				if prefix, pOk := binding["prefix"]; pOk {
					p := prefix.(string)
					out.Prefix = &(p)
				}
			}
		}
		a, err := NewDataType(base["items"])
		if err != nil {
			return out, fmt.Errorf("Unable to parse type: %s", err)
		}
		out.Items = &a
		return out, nil
	} else {
		return DataType{}, fmt.Errorf("Unknown data type: %#v\n", value)
	}
	return DataType{}, nil
}

func NewArgument(x interface{}) (Argument, error) {
	if base, ok := x.(string); ok {
		return Argument{Value: &base}, nil
	}
	if base, ok := x.(map[interface{}]interface{}); ok {
		out := Argument{}
		if x, ok := base["valueFrom"]; ok {
			s := x.(string)
			out.ValueFrom = &s
		}
		if x, ok := base["position"]; ok {
			out.Position = x.(int)
		} else {
			out.Position = 10000
		}
		if x, ok := base["prefix"]; ok {
			x_s := x.(string)
			out.Prefix = &x_s
		}
		return out, nil
	}
	return Argument{}, fmt.Errorf("Can't Parse Argument")
}

func NewRequirements(x interface{}) ([]Requirement, error) {
	out := []Requirement{}
	if base, ok := x.([]interface{}); ok {
		for _, i := range base {
			o, err := NewRequirement(i)
			if err != nil {
				return out, err
			}
			out = append(out, o)
		}
	} else {
		return out, fmt.Errorf("Unable to parse requirements block")
	}
	return out, nil
}

func NewRequirement(x interface{}) (Requirement, error) {
	if base, ok := x.(map[interface{}]interface{}); ok {
		if id, ok := base["class"]; ok {
			id_string := id.(string)
			switch {
			case id_string == "SchemaDefRequirement":
				return NewSchemaDefRequirement(base)
			default:
				return nil, fmt.Errorf("Unknown requirement: %s", id_string)
			}
		} else {
			return nil, fmt.Errorf("Undefined requirement")
		}
	}
	return nil, fmt.Errorf("Undefined requirement")
}

func NewSchemaDefRequirement(x map[interface{}]interface{}) (SchemaDefRequirement, error) {
	newTypes := []DataType{}
	if base, ok := x["types"]; ok {
		if fieldArray, ok := base.([]interface{}); ok {
			for _, i := range fieldArray {
				d, err := NewDataType(i)
				if err != nil {
					return SchemaDefRequirement{}, fmt.Errorf("Unknown DataType: %s", err)
				}
				newTypes = append(newTypes, d)
			}
		}
	} else {
		return SchemaDefRequirement{}, fmt.Errorf("No types column")
	}
	return SchemaDefRequirement{NewTypes: newTypes}, nil
}
