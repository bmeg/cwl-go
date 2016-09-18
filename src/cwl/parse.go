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
	"strings"
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
			} else if class == "Directory" {
				for k, v := range base {
					if k == "path" {
						out["path"] = filepath.Join(basePath, v.(string))
					} else if k == "location" {
						out["location"] = filepath.Join(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			} else {
				log.Printf("Unknown class type: %s", class)
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

func Parse(cwl_path string) (CWLGraph, error) {
	source, err := ioutil.ReadFile(cwl_path)
	if err != nil {
		return CWLGraph{}, fmt.Errorf("Unable to parse file: %s", err)
	}
	doc := make(map[interface{}]interface{})
	err = yaml.Unmarshal(source, &doc)
	x, _ := filepath.Abs(cwl_path)
	parser := CWLParser{Path: x, Schemas: make(map[string]Schema), Elements: make(map[string]CWLDoc)}
	if base, ok := doc["$graph"]; ok {
		return parser.NewGraph(base)
	} else if _, ok := doc["class"]; ok {
		return parser.NewClass(doc)
	}
	return CWLGraph{}, fmt.Errorf("Unable to parse file")
}

type CWLParser struct {
	Path     string
	Schemas  map[string]Schema
	Elements map[string]CWLDoc
}

func (self *CWLParser) AddSchema(schema Schema) {
	self.Schemas[schema.Name] = schema
}

func (self *CWLParser) GetElement(path string) (CWLDoc, error) {
	if i, ok := self.Elements[path]; ok {
		return i, nil
	} else {
		if strings.HasPrefix(path, "#") {
			log.Printf("Need to parse another part of the graph")
			source, err := ioutil.ReadFile(self.Path)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse file: %s", err)
			}
			doc := make(map[interface{}]interface{})
			yaml.Unmarshal(source, &doc)
			parser := CWLParser{Path: self.Path, Schemas: make(map[string]Schema), Elements: make(map[string]CWLDoc)}
			if _, ok := doc["$graph"]; ok {
				p := path[1:]
				if base, ok := doc["$graph"].([]interface{}); ok {
					for _, i := range base {
						if bmap, ok := i.(map[interface{}]interface{}); ok {
							if bmap["id"] == p {
								log.Printf("Found it")
								c, err := parser.NewClass(bmap)
								if err != nil {
									return nil, err
								}
								d := c.Elements[c.Main]
								self.Elements[path] = d
								return d, nil
							}
						}
					}
				}
			} else {
				return nil, fmt.Errorf("Not a cwl graph file")
			}
		} else {
			script := filepath.Join(filepath.Dir(self.Path), path)
			cDoc, err := Parse(script)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse script %s", script)
			}
			for k, v := range cDoc.Elements {
				self.Elements[k] = v
			}
		}
	}
	return nil, fmt.Errorf("Unable to parse script")
}

func (self *CWLParser) NewClass(doc map[interface{}]interface{}) (CWLGraph, error) {
	if doc["class"].(string) == "Workflow" {
		return self.NewWorkflow(doc)
	} else if doc["class"].(string) == "CommandLineTool" {
		return self.NewCommandLineTool(doc)
	} else if doc["class"].(string) == "ExpressionTool" {
		return self.NewExpressionTool(doc)
	}
	return CWLGraph{}, fmt.Errorf("Unknown class type")
}

func (self *CWLParser) NewGraph(graph interface{}) (CWLGraph, error) {
	docs := CWLGraph{Elements: map[string]CWLDoc{}}
	if base, ok := graph.([]interface{}); ok {
		for _, i := range base {
			parser := CWLParser{Path: self.Path, Schemas: make(map[string]Schema), Elements: make(map[string]CWLDoc)}
			if classBase, ok := i.(map[interface{}]interface{}); ok {
				cDoc, err := parser.NewClass(classBase)
				if err != nil {
					return docs, err
				}
				for k, v := range cDoc.Elements {
					docs.Elements[k] = v
				}
				log.Printf("Parsing Graph %#v", i)
			}
		}
	}
	return docs, nil
}

func (self *CWLParser) NewWorkflow(doc map[interface{}]interface{}) (CWLGraph, error) {
	log.Printf("Workflow: %v", doc)
	out := Workflow{}
	if _, ok := doc["id"]; ok {
		out.Id = doc["id"].(string)
	} else {
		out.Id = "#_main"
	}

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
					return CWLGraph{}, fmt.Errorf("Workflow Input error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewWorkflowInput("", x)
				if err == nil {
					out.Inputs[n.Id] = n
				} else {
					return CWLGraph{}, fmt.Errorf("Workflow Input error: %s", err)
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
					return CWLGraph{}, fmt.Errorf("Workflow Output error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewWorkflowOutput("", x)
				if err == nil {
					out.Outputs[n.Id] = n
				} else {
					return CWLGraph{}, fmt.Errorf("Workflow Output error: %s", err)
				}
			}
		}
	}

	if base, ok := doc["steps"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewStep(k.(string), v)
				if err == nil {
					n.Parent = &out
					out.Steps[n.Id] = n
				} else {
					return CWLGraph{}, fmt.Errorf("Workflow Step error: %s", err)
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			for _, x := range base_array {
				n, err := self.NewStep("", x)
				if err == nil {
					n.Parent = &out
					out.Steps[n.Id] = n
				} else {
					return CWLGraph{}, fmt.Errorf("Workflow Step error: %s", err)
				}
			}
		}
	}
	return CWLGraph{Elements: map[string]CWLDoc{out.Id: out}, Main: out.Id}, nil
}

func (self *CWLParser) NewCommandLineTool(doc map[interface{}]interface{}) (CWLGraph, error) {
	log.Printf("CommandLineTool: %v", doc)
	out := CommandLineTool{}
	out.Inputs = make(map[string]CommandInput)
	out.Outputs = make(map[string]CommandOutput)

	if _, ok := doc["id"]; ok {
		out.Id = doc["id"].(string)
	} else {
		out.Id = "#_main"
	}

	/* Requirements */
	if base, ok := doc["requirements"]; ok {
		r, err := self.NewRequirements(base)
		if err != nil {
			return CWLGraph{}, err
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
			} else {
				return CWLGraph{}, fmt.Errorf("Error Parsing Arguments: %s", err)
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
					return CWLGraph{}, fmt.Errorf("Command line Input error: %s", err)
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
					return CWLGraph{}, fmt.Errorf("Output map parsing errors: %s %s %s %#v", out.Id, k, err, v)
				}
				out.Outputs[n.Id] = n
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Output array: %d", len(base_array))
			for _, x := range base_array {
				n, err := self.NewCommandOutput("", x)
				if err != nil {
					return CWLGraph{}, fmt.Errorf("Output array parsing errors: %s %s %#v", n.Id, err, x)
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

	if base, ok := doc["successCodes"]; ok {
		out.SuccessCodes = []int{}
		if abase, ok := base.([]interface{}); ok {
			for _, i := range abase {
				out.SuccessCodes = append(out.SuccessCodes, i.(int))
			}
		}
	}

	log.Printf("Parse CommandLineTool: %v", out)
	return CWLGraph{Elements: map[string]CWLDoc{out.Id: out}, Main: out.Id}, nil
}

func (self *CWLParser) NewExpressionTool(doc map[interface{}]interface{}) (CWLGraph, error) {
	log.Printf("ExpressionTool: %v", doc)
	out := ExpressionTool{}
	out.Inputs = make(map[string]ExpressionInput)
	out.Outputs = make(map[string]ExpressionOutput)

	if _, ok := doc["id"]; ok {
		out.Id = doc["id"].(string)
	} else {
		out.Id = "#_main"
	}

	/* Requirements */
	if base, ok := doc["requirements"]; ok {
		r, err := self.NewRequirements(base)
		if err != nil {
			return CWLGraph{}, err
		}
		log.Printf("Requirements: %#v", r)
		out.Requirements = r
	}

	if base, ok := doc["expression"].(string); ok {
		out.Expression = base
	}

	/* Inputs */
	if base, ok := doc["inputs"]; ok {
		if base_map, ok := base.(map[interface{}]interface{}); ok {
			for k, v := range base_map {
				n, err := self.NewExpressionInput(k.(string), v)
				if err == nil {
					out.Inputs[n.Id] = n
				}
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Input array: %d", len(base_array))
			for _, x := range base_array {
				n, err := self.NewExpressionInput("", x)
				if err == nil {
					out.Inputs[n.Id] = n
				} else {
					return CWLGraph{}, fmt.Errorf("Command line Input error: %s", err)
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
				n, err := self.NewExpressionOutput(k.(string), v)
				if err != nil {
					return CWLGraph{}, fmt.Errorf("Output map parsing errors: %s %s %s %#v", out.Id, k, err, v)
				}
				out.Outputs[n.Id] = n
			}
		} else if base_array, ok := base.([]interface{}); ok {
			log.Printf("Output array: %d", len(base_array))
			for _, x := range base_array {
				n, err := self.NewExpressionOutput("", x)
				if err != nil {
					return CWLGraph{}, fmt.Errorf("Output array parsing errors: %s %s %#v", n.Id, err, x)
				}
				out.Outputs[n.Id] = n
			}
		} else {
			log.Printf("Can't Parse Outputs")
		}
	} else {
		log.Printf("No Outputs found")
	}

	log.Printf("Parse ExpressionTool: %v", out)
	return CWLGraph{Elements: map[string]CWLDoc{out.Id: out}, Main: out.Id}, nil
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
		if s, ok := base["source"]; ok {
			out.OutputSource = s.(string)
		}
		if s, ok := base["outputSource"]; ok {
			out.OutputSource = s.(string)
		}
	}
	return out, nil
}

func (self *CWLParser) NewStep(id string, x interface{}) (Step, error) {
	sout := Step{}
	sout.In = map[string]StepInput{}
	sout.Out = map[string]StepOutput{}

	if base, ok := x.(map[interface{}]interface{}); ok {
		if id == "" {
			sout.Id = base["id"].(string)
		} else {
			sout.Id = id
		}

		if bIn, ok := base["in"]; ok {
			inputs, err := self.NewStepInputSet(bIn)
			if err != nil {
				return sout, err
			}
			sout.In = inputs
		} else if bIn, ok := base["inputs"]; ok {
			inputs, err := self.NewStepInputSet(bIn)
			if err != nil {
				return sout, err
			}
			sout.In = inputs
		} else {
			log.Printf("Step %s has no inputs", sout.Id)
		}

		if bOut, ok := base["out"]; ok {
			outputs, err := self.NewStepOutputSet(bOut)
			if err != nil {
				return sout, err
			}
			sout.Out = outputs
		} else if bOut, ok := base["outputs"]; ok {
			outputs, err := self.NewStepOutputSet(bOut)
			if err != nil {
				return sout, err
			}
			sout.Out = outputs
		} else {
			log.Printf("Step %s has no output", sout.Id)
		}

		if bRun, ok := base["run"]; ok {
			r := bRun.(string)
			log.Printf("StepRun: %s", r)
			doc, err := self.GetElement(r)
			if err != nil {
				return sout, fmt.Errorf("Unable to parse step: %s", err)
			}
			sout.Doc = doc
		}

	} else {
		return sout, fmt.Errorf("Unable to parse step")
	}
	return sout, nil
}

func (self *CWLParser) NewStepInputSet(x interface{}) (map[string]StepInput, error) {
	sOut := map[string]StepInput{}
	if in, ok := x.(map[interface{}]interface{}); ok {
		for k, v := range in {
			i, err := self.NewStepInput(k.(string), v)
			if err != nil {
				return sOut, fmt.Errorf("Unable to parse step input element: %s", err)
			}
			sOut[i.Id] = i
		}
	} else if in, ok := x.([]interface{}); ok {
		for _, v := range in {
			i, err := self.NewStepInput("", v)
			if err != nil {
				return sOut, fmt.Errorf("Unable to parse step input element: %s", err)
			}
			sOut[i.Id] = i
		}
	} else {
		return sOut, fmt.Errorf("Unable to parse step input set")
	}
	return sOut, nil
}

func (self *CWLParser) NewStepOutputSet(x interface{}) (map[string]StepOutput, error) {
	sOut := map[string]StepOutput{}

	if out, ok := x.(map[interface{}]interface{}); ok {
		for k, v := range out {
			i, err := self.NewStepOutput(k.(string), v)
			if err != nil {
				return sOut, fmt.Errorf("Unable to parse step output element: %s", err)
			}
			sOut[k.(string)] = i
		}
	} else if out, ok := x.([]interface{}); ok {
		for _, v := range out {
			i, err := self.NewStepOutput("", v)
			if err != nil {
				return sOut, fmt.Errorf("Unable to parse step output element: %s", err)
			}
			sOut[i.Id] = i
		}
	} else {
		return sOut, fmt.Errorf("Unable to parse step output set: %#v", x)
	}
	return sOut, nil
}

func (self *CWLParser) NewStepInput(id string, x interface{}) (StepInput, error) {
	out := StepInput{}
	if id != "" {
		out.Id = id
	}

	if base, ok := x.(map[interface{}]interface{}); ok {
		if id, ok := base["id"]; ok {
			out.Id = id.(string)
		}

		if source, ok := base["source"]; ok {
			out.Source = source.(string)
		}
		if defaultVal, ok := base["default"]; ok {
			out.Default = &defaultVal
		}
	} else if base, ok := x.(string); ok {
		out.Source = base
	} else {
		return out, fmt.Errorf("Unable to parse step input")
	}

	return out, nil
}

func (self *CWLParser) NewStepOutput(id string, x interface{}) (StepOutput, error) {
	out := StepOutput{}
	if id != "" {
		out.Id = id
	}

	if base, ok := x.(string); ok {
		out.Id = base
	} else if base, ok := x.(map[interface{}]interface{}); ok {
		if b, ok := base["id"]; ok {
			out.Id = b.(string)
		}
	} else {
		return out, fmt.Errorf("Unable to parse step input")
	}

	return out, nil
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
		} else if bType, ok := base["type"]; ok {
			if bType == "stdout" {
				//do nothing here, just avoiding warning logging
			} else if bType == "stderr" {
				//do nothing here, just avoiding warning logging
			} else {
				log.Printf("Unknown output binding: %s %s", out.Id, base)
			}
		} else {
			log.Printf("No output binding: %s %s", out.Id, base)
		}
	} else if base, ok := x.(string); ok {
		log.Printf("output schema string: %s", base)
	} else {
		return out, fmt.Errorf("Unable to parse CommandOutput: %v", x)
	}
	return out, nil
}

func (self *CWLParser) NewExpressionInput(id string, x interface{}) (ExpressionInput, error) {
	log.Printf("ExpressionInput parse")
	t, err := self.NewSchema(x)
	if err != nil {
		return ExpressionInput{}, fmt.Errorf("unable to load schema: %s", err)
	}
	out := ExpressionInput{Schema: t}
	if id != "" {
		out.Id = id
	}
	return out, nil
}

func (self *CWLParser) NewExpressionOutput(id string, x interface{}) (ExpressionOutput, error) {
	log.Printf("ExpressionOutput parse")
	t, err := self.NewSchema(x)
	if err != nil {
		return ExpressionOutput{}, fmt.Errorf("unable to load schema: %s", err)
	}
	out := ExpressionOutput{Schema: t}
	if id != "" {
		out.Id = id
	}
	return out, nil
}

var SCHEMA_TYPES = map[string]bool{
	"boolean":   true,
	"int":       true,
	"array":     true,
	"record":    true,
	"File":      true,
	"Directory": true,
	"null":      true,
	"string":    true,
	"stdout":    true,
	"stderr":    true,
	"Any":       true,
}

func (self *CWLParser) NewSchema(value interface{}) (Schema, error) {

	if base, ok := value.(string); ok {
		if _, found := SCHEMA_TYPES[base]; !found {
			log.Printf("Schema not found: %s", base)
			if _, ok := self.Schemas[base[1:]]; ok {
				log.Printf("Schema Found")
			} else {
				log.Printf("Not found in %#v", self.Schemas)
			}
			return Schema{}, fmt.Errorf("Schema not found: %s", base)
		} else {
			return Schema{TypeName: base}, nil
		}
	}

	if base, ok := value.(map[interface{}]interface{}); ok {
		out := Schema{}
		if tname, ok := base["type"].(string); ok {
			o, err := self.NewSchema(tname)
			if err != nil {
				return out, err
			}
			out = o
		} else if tstruct, ok := base["type"].(map[interface{}]interface{}); ok {
			o, err := self.NewSchema(tstruct)
			if err != nil {
				return out, fmt.Errorf("Unable to parse type schema: %s: %s", tstruct, err)
			}
			out.Types = []Schema{o}
			out.TypeName = "array_holder"
		} else if tarray, ok := base["type"].([]interface{}); ok {
			for _, i := range tarray {
				a, err := self.NewSchema(i)
				if err != nil {
					return out, fmt.Errorf("Unable to parse type element: %s", err)
				}
				out.Types = append(out.Types, a)
			}
		} else {
			return out, fmt.Errorf("Can't parse type: %#v", base["type"])
		}

		if id, ok := base["id"]; ok {
			out.Id = id.(string)
		}
		if name, ok := base["name"]; ok {
			out.Name = name.(string)
		}

		if binding, ok := base["inputBinding"]; ok {
			out.Bound = true
			if pos, ok := binding.(map[interface{}]interface{})["position"]; ok {
				out.Position = pos.(int)
			} else {
				out.Position = 100000
			}
			if prefix, ok := binding.(map[interface{}]interface{})["prefix"].(string); ok {
				out.Prefix = prefix
			}
			if itemSep, ok := binding.(map[interface{}]interface{})["itemSeparator"].(string); ok {
				out.ItemSeparator = itemSep
			}
		}

		if def, ok := base["default"]; ok {
			out.Default = &def
			//special case when default value is a file
			if out.TypeName == "File" {
				if base, ok := def.(map[interface{}]interface{}); ok {
					if s, ok := base["path"]; ok {
						base["path"] = path.Join(filepath.Dir(self.Path), s.(string))
					}
					if s, ok := base["location"]; ok {
						base["location"] = path.Join(filepath.Dir(self.Path), s.(string))
					}
				}
			}
		}

		if bItem, ok := base["items"]; ok {
			a, err := self.NewSchema(bItem)
			if err != nil {
				return out, fmt.Errorf("Can't parse items")
			}
			log.Printf("Items Schema: %#v", a)
			out.Items = &a
		}
		log.Printf("NewSchema: %#v", out)
		return out, nil
	} else {
		return Schema{}, fmt.Errorf("Unknown data type: %#v", value)
	}
	return Schema{}, nil
}

func (self *CWLParser) NewArgument(x interface{}) (Argument, error) {
	if base, ok := x.(string); ok {
		return Argument{Value: &base, Schema: Schema{Bound: true}}, nil
	}
	if base, ok := x.(map[interface{}]interface{}); ok {
		out := Argument{Schema: Schema{Bound: true}}
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
			out.Prefix = x_s
		}
		return out, nil
	}
	return Argument{}, fmt.Errorf("Can't Parse Argument")
}

func (self *CWLParser) NewRequirements(x interface{}) ([]Requirement, error) {
	out := []Requirement{}
	if base, ok := x.([]interface{}); ok {
		for _, i := range base {
			if base, ok := i.(map[interface{}]interface{}); ok {
				if id, ok := base["class"]; ok {
					id_string := id.(string)
					o, err := self.NewRequirement(id_string, i)
					if err != nil {
						return out, err
					}
					out = append(out, o)
				}
			}
		}
	} else if base, ok := x.(map[interface{}]interface{}); ok {
		for k, v := range base {
			o, err := self.NewRequirement(k.(string), v)
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

func (self *CWLParser) NewRequirement(id_string string, conf interface{}) (Requirement, error) {
	log.Printf("Requirement: %s", id_string)
	switch {
	case id_string == "SchemaDefRequirement":
		schemaRequirement, err := self.NewSchemaDefRequirement(conf)
		if err != nil {
			return schemaRequirement, err
		}
		for _, i := range schemaRequirement.NewTypes {
			self.AddSchema(i)
		}
		return schemaRequirement, nil
	case id_string == "InlineJavascriptRequirement":
		return self.NewInlineJavascriptRequirement(conf)
	case id_string == "InitialWorkDirRequirement":
		return self.NewInitialWorkDirRequirement(conf)
	default:
		log.Printf("Unsupported Requirement %s", id_string)
		e := UnsupportedRequirement{Message: fmt.Sprintf("Unknown requirement: %s", id_string)}
		return nil, e
	}
	return nil, fmt.Errorf("Undefined requirement")
}

func (self *CWLParser) NewSchemaDefRequirement(conf interface{}) (SchemaDefRequirement, error) {
	newTypes := []Schema{}
	if x, ok := conf.(map[interface{}]interface{}); ok {
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
	}
	return SchemaDefRequirement{NewTypes: newTypes}, nil
}

func (self *CWLParser) NewInlineJavascriptRequirement(x interface{}) (InlineJavascriptRequirement, error) {
	return InlineJavascriptRequirement{}, nil
}

func (self *CWLParser) NewInitialWorkDirRequirement(x interface{}) (InitialWorkDirRequirement, error) {
	return InitialWorkDirRequirement{}, nil
}
