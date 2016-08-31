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
	doc := make(map[interface{}]interface{})
	err = yaml.Unmarshal(source, &doc)
	x, _ := filepath.Abs(path)
	base_path := filepath.Dir(x)

	out := AdjustInputs(doc, base_path).(map[interface{}]interface{})
	log.Printf("Inputs: %#v", out)
	return out, err
}

func AdjustInputs(input interface{}, basePath string) interface{} {
	if base, ok := input.(map[interface{}]interface{}); ok {
		out := map[interface{}]interface{}{}
		if class, ok := base["class"]; ok {
			log.Printf("class: %s", class)
			if class == "File" {
				for k, v := range base {
					if k == "path" {
						out["path"] = filepath.Join(basePath, v.(string))
					} else if k == "location" {
						out["location"] = filepath.Join(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			}
		} else {
			for k, v := range base {
				out[k] = AdjustInputs(v, basePath)
			}
		}
		return out
	} else if base, ok := input.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, AdjustInputs(i, basePath))
		}
		return out
	}
	return input
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
	parser := CWLParser{BasePath: base_path}
	if doc["class"].(string) == "Workflow" {
		return parser.NewWorkflow(doc)
	}
	if doc["class"].(string) == "CommandLineTool" {
		return parser.NewCommandLineTool(doc)
	}
	return nil, nil
}

type CWLParser struct {
	BasePath string
}

func (self *CWLParser) NewWorkflow(doc CWLDocData) (CWLDoc, error) {
	log.Printf("Workflow: %v", doc)
	out := Workflow{}
	out.Inputs = make(map[string]WorkflowInput)
	out.Outputs = make(map[string]WorkflowOutput)

	out.Steps = make(map[string]Step)

	if base, ok := doc["inputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewWorkflowInput(k.(string), v)
				if err == nil {
					out.Inputs[n.Id] = n
				} else {
					log.Printf("Workflow Input error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewWorkflowInput("", x)
				if err == nil {
					out.Inputs[n.Id] = n
				} else {
					log.Printf("Workflow Input error: %s", err)
				}
			}
		}
	}

	if base, ok := doc["outputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewWorkflowOutput(k.(string), v)
				if err == nil {
					out.Outputs[n.Id] = n
				} else {
					log.Printf("Workflow Output error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewWorkflowOutput("", x)
				if err == nil {
					out.Outputs[n.Id] = n
				} else {
					log.Printf("Workflow Output error: %s", err)
				}
			}
		}
	}

	if base, ok := doc["steps"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewStep(k.(string), v)
				if err == nil {
					out.Steps[n.Id] = n
				} else {
					log.Printf("Workflow Step error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewStep("", x)
				if err == nil {
					out.Steps[n.Id] = n
				} else {
					log.Printf("Workflow Step error: %s", err)
				}
			}
		}
	}
	return out, nil
}

func (self *CWLParser) NewCommandLineTool(doc CWLDocData) (CWLDoc, error) {
	log.Printf("CommandLineTool: %v", doc)
	out := CommandLineTool{}
	out.Inputs = make(map[string]CommandInput)
	out.Outputs = make(map[string]CommandOutput)

	if _, ok := doc["id"]; ok {
		out.Id = doc["id"].(string)
	} else {
		out.Id = ""
	}

	/* Requirements */
	if base, ok := doc["requirements"]; ok {
		r, err := self.NewRequirements(base)
		if err != nil {
			return CommandLineTool{}, err
		}
		log.Printf("Requirements: %#v", r)
		out.Requirements = r
	}

	/* BaseCommand */
	if base, ok := doc["baseCommand"].([]interface{}); ok {
		o := make([]string, len(base))
		for i, v := range base {
			o[i] = v.(string)
		}
		out.BaseCommand = o
	} else {
		if base, ok := doc["baseCommand"].(string); ok {
			out.BaseCommand = []string{base}
		}
	}

	/* Arguments */
	if base, ok := doc["arguments"]; ok {
		for _, x := range base.([]interface{}) {
			n, err := self.NewArgument(x)
			if err == nil {
				out.Arguments = append(out.Arguments, n)
			}
		}
	}

	/* Inputs */
	if base, ok := doc["inputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewCommandInput(k.(string), v)
				if err == nil {
					out.Inputs[n.Id] = n
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Input array: %d", len(base_array))
			for _, x := range base_array {
				n, err := self.NewCommandInput("", x)
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

	/* Outputs */
	if base, ok := doc["outputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewCommandOutput(k.(string), v)
				if err != nil {
					return out, fmt.Errorf("Output map parsing errors: %s %s %s %#v", out.Id, k, err, v)
				}
				out.Outputs[n.Id] = n
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Output array: %d", len(base_array))
			for _, x := range base_array {
				n, err := self.NewCommandOutput("", x)
				if err != nil {
					return out, fmt.Errorf("Output array parsing errors: %s %s", out.Id, err)
				}
				out.Outputs[n.Id] = n
			}
		} else {
			log.Printf("Can't Parse Outputs")
		}
	} else {
		log.Printf("No Outputs found")
	}

	if base, ok := doc["stderr"]; ok {
		out.Stderr = base.(string)
	}
	if base, ok := doc["stdout"]; ok {
		out.Stdout = base.(string)
	}
	if base, ok := doc["stdin"]; ok {
		out.Stdin = base.(string)
	}

	log.Printf("Parse CommandLineTool: %v", out)
	return out, nil
}

func (self *CWLParser) NewWorkflowInput(id string, x interface{}) (WorkflowInput, error) {
	t, err := self.NewSchema(x)
	if err != nil {
		return WorkflowInput{}, fmt.Errorf("unable to load data type: %s", err)
	}
	if id != "" {
		t.Id = id
	}
	out := WorkflowInput{Schema: t}
	return out, nil
}

func (self *CWLParser) NewWorkflowOutput(id string, x interface{}) (WorkflowOutput, error) {
	t, err := self.NewSchema(x)
	if err != nil {
		return WorkflowOutput{}, fmt.Errorf("unable to load data type: %s", err)
	}
	if id != "" {
		t.Id = id
	}
	out := WorkflowOutput{Schema: t}
	if base, ok := x.(map[interface{}]interface{}); ok {
		if s, ok := base["outputSource"]; ok {
			out.OutputSource = s.(string)
		}
	}
	return out, nil
}

func (self *CWLParser) NewStep(id string, x interface{}) (Step, error) {
	sout := Step{}
	sout.In = map[string]string{}
	sout.Out = map[string]string{}

	if base, ok := x.(map[interface{}]interface{}); ok {
		if id == "" {
			sout.Id = base["id"].(string)
		} else {
			sout.Id = id
		}

		if bIn, ok := base["in"]; ok {
			if in, ok := bIn.(map[interface{}]interface{}); ok {
				for k, v := range in {
					sout.In[k.(string)] = v.(string)
				}
			}
		}

		if bOut, ok := base["out"]; ok {
			if out, ok := bOut.(map[interface{}]interface{}); ok {
				for k, v := range out {
					sout.Out[k.(string)] = v.(string)
				}
			}
		}

		if bRun, ok := base["run"]; ok {
			r := bRun.(string)
			script := filepath.Join(self.BasePath, r)
			log.Printf("RunScript: %s", script)
			cDoc, err := Parse(script)
			if err != nil {
				return sout, fmt.Errorf("Unable to parse script %s", script)
			}
			sout.Doc = cDoc
		}

	}
	return sout, nil
}

func (self *CWLParser) NewCommandInput(id string, x interface{}) (CommandInput, error) {
	t, err := self.NewSchema(x)
	if err != nil {
		return CommandInput{}, fmt.Errorf("unable to load data type: %s", err)
	}
	out := CommandInput{Schema: t}
	if id != "" {
		out.Id = id
	}
	return out, nil
}

func (self *CWLParser) NewCommandOutput(id string, x interface{}) (CommandOutput, error) {
	log.Printf("CommandOutput parse")
	t, err := self.NewSchema(x)
	if err != nil {
		return CommandOutput{}, fmt.Errorf("unable to load schema: %s", err)
	}
	out := CommandOutput{Schema: t}
	if id != "" {
		out.Id = id
	}

	if base, ok := x.(map[interface{}]interface{}); ok {
		if _, ok := base["outputBinding"]; ok {
			if bindBase, ok := base["outputBinding"].(map[interface{}]interface{}); ok {
				if _, ok := bindBase["glob"]; ok {
					g := bindBase["glob"].(string)
					out.Glob = g
				}
			} else {
				log.Printf("Output Binding format weird")
			}
		} else {
			log.Printf("No output binding: %s", out.Id)
		}
	} else {
		return out, fmt.Errorf("Unable to parse CommandOutput: %v", x)
	}
	return out, nil
}

func (self *CWLParser) NewSchema(value interface{}) (Schema, error) {

	if base, ok := value.(map[interface{}]interface{}); ok {
		out := Schema{TypeName: base["type"].(string)}
		if id, ok := base["id"]; ok {
			out.Id = id.(string)
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

			if def, ok := base["default"]; ok {
				out.Default = &def
				//special case when default value is a file
				if out.TypeName == "File" {
					if base, ok := def.(map[interface{}]interface{}); ok {
						if s, ok := base["path"]; ok {
							base["path"] = path.Join(self.BasePath, s.(string))
						}
						if s, ok := base["location"]; ok {
							base["location"] = path.Join(self.BasePath, s.(string))
						}
					}
				}
			}
		}
		if _, ok := base["items"]; ok {
			a, err := self.NewSchema(base["items"])
			if err != nil {
				return out, fmt.Errorf("Unable to parse type: %s", err)
			}
			out.Items = &a
		}
		return out, nil
	} else {
		return Schema{}, fmt.Errorf("Unknown data type: %#v\n", value)
	}
	return Schema{}, nil
}

func (self *CWLParser) NewArgument(x interface{}) (Argument, error) {
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

func (self *CWLParser) NewRequirements(x interface{}) ([]Requirement, error) {
	out := []Requirement{}
	if base, ok := x.([]interface{}); ok {
		for _, i := range base {
			o, err := self.NewRequirement(i)
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

func (self *CWLParser) NewRequirement(x interface{}) (Requirement, error) {
	if base, ok := x.(map[interface{}]interface{}); ok {
		if id, ok := base["class"]; ok {
			id_string := id.(string)
			switch {
			case id_string == "SchemaDefRequirement":
				return self.NewSchemaDefRequirement(base)
			case id_string == "InlineJavascriptRequirement":
				return self.NewInlineJavascriptRequirement(base)
			case id_string == "InitialWorkDirRequirement":
				return self.NewInitialWorkDirRequirement(base)
			default:
				return nil, fmt.Errorf("Unknown requirement: %s", id_string)
			}
		} else {
			return nil, fmt.Errorf("Undefined requirement")
		}
	}
	return nil, fmt.Errorf("Undefined requirement")
}

func (self *CWLParser) NewSchemaDefRequirement(x map[interface{}]interface{}) (SchemaDefRequirement, error) {
	newTypes := []Schema{}
	if base, ok := x["types"]; ok {
		if fieldArray, ok := base.([]interface{}); ok {
			for _, i := range fieldArray {
				d, err := self.NewSchema(i)
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

func (self *CWLParser) NewInlineJavascriptRequirement(x map[interface{}]interface{}) (InlineJavascriptRequirement, error) {
	return InlineJavascriptRequirement{}, nil
}

func (self *CWLParser) NewInitialWorkDirRequirement(x map[interface{}]interface{}) (InitialWorkDirRequirement, error) {
	return InitialWorkDirRequirement{}, nil
}
