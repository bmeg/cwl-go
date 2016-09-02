package cwl

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

func (self CommandLineTool) NewGraphState(inputs JSONDict) GraphState {
	return GraphState{INPUT_FIELD: JobState{RESULTS_FIELD: inputs}}
}

func (self CommandLineTool) GetIDs() []string {
	return []string{self.Id}
}

func (self CommandLineTool) Done(state GraphState) bool {
	if _, ok := state[self.Id]; ok {
		return true
	}
	return false
}

func (self CommandLineTool) UpdateStepResults(state GraphState, stepId string, results JSONDict) GraphState {
	out := GraphState{}
	for k, v := range state {
		out[k] = v
	}
	out[stepId] = JobState{RESULTS_FIELD: results}
	return out
}

func (self CommandLineTool) ReadySteps(state GraphState) []string {
	if _, ok := state[self.Id]; ok {
		return []string{}
	} else if _, ok := state["#"]; ok {
		return []string{self.Id}
	}
	return []string{}
}

func (self CommandLineTool) GetResults(state GraphState) JSONDict {
	return state[self.Id][RESULTS_FIELD].(JSONDict)
}

func (self CommandLineTool) GenerateJob(step string, graphState GraphState) (Job, error) {
	if i, ok := graphState[INPUT_FIELD]; !ok {
		return Job{}, fmt.Errorf("%s Inputs not ready", step)
	} else {
		cmd, files, err := self.Evaluate(i[RESULTS_FIELD].(JSONDict))
		if err != nil {
			log.Printf("Job Eval Error: %s", err)
			return Job{}, err
		}
		return Job{Cmd: cmd, Files: files, Stderr: self.Stderr, Stdout: self.Stdout, Stdin: self.Stdin, Inputs: i[RESULTS_FIELD].(JSONDict)}, nil
	}
}

type cmdArgArray []cmdArg

func (self cmdArgArray) Len() int {
	return len(self)
}

func (self cmdArgArray) Less(i, j int) bool {
	if (self)[i].position == (self)[j].position {
		return (self)[i].id < (self)[j].id
	}
	return (self)[i].position < (self)[j].position
}

func (self cmdArgArray) Swap(i, j int) {
	(self)[i], (self)[j] = (self)[j], (self)[i]
}

func (self *CommandLineTool) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	log.Printf("CommandLineTool Evalute")
	out := make([]string, 0)
	out = append(out, self.BaseCommand...)

	oFiles := []JobFile{}

	args := make(cmdArgArray, 0, len(self.Arguments)+len(self.Inputs))
	//Arguments
	for _, x := range self.Arguments {
		e, files, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Argument Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		c := cmdArg{
			position: x.Position,
			id:       "",
			value:    e,
		}
		args = append(args, c)
		oFiles = append(oFiles, files...)
	}
	//Inputs
	for _, x := range self.Inputs {
		e, files, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Input Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		c := cmdArg{
			position: x.Position,
			id:       x.Id,
			value:    e,
		}
		args = append(args, c)
		oFiles = append(oFiles, files...)
	}

	//Outputs
	for _, x := range self.Outputs {
		_, files, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Output Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		oFiles = append(oFiles, files...)
	}

	sort.Stable(args)
	for _, x := range args {
		out = append(out, x.value...)
	}
	log.Printf("Out: %v", out)
	return out, oFiles, nil
}

func (self *Schema) IsOptional() bool {
	for _, a := range self.Types {
		if a.TypeName == "null" {
			return true
		}
	}
	return false
}

func (self *Schema) SchemaEvaluate(value interface{}) ([]string, []JobFile, error) {
	out_args := []string{}
	out_files := []JobFile{}

	typeName := self.TypeName

	if typeName != "array_holder" {
		for _, a := range self.Types {
			if a.TypeName == "null" && value == nil {
				typeName = "null"
			} else {
				//BUG: this is assuming a binary choice null vs something...
				typeName = a.TypeName
			}
		}
	}

	if typeName == "File" {
		if base, ok := value.(map[interface{}]interface{}); ok {
			if class, ok := base["class"]; ok {
				if class.(string) == "File" {
					loc := base["location"].(string)
					out_args = []string{loc}
					out_files = []JobFile{JobFile{Location: loc}}
				} else {
					log.Printf("Unknown class %s", class)
				}
			} else {
				log.Printf("Input map has no class")
			}
		} else {
			log.Printf("File input not formatted correctly: %#v", value)
		}
	} else if typeName == "int" {
		out_args = []string{fmt.Sprintf("%d", value.(int))}
	} else if typeName == "boolean" {
		if value.(bool) {
			out_args = []string{self.Prefix}
		}
	} else if typeName == "array_holder" {
		o, f, err := self.Types[0].SchemaEvaluate(value)
		if err != nil {
			return []string{}, []JobFile{}, fmt.Errorf("Bad array '%s' (%#v): %s", typeName, *self, err)
		}
		out_args = o
		out_files = f
	} else if typeName == "array" {
		if base, ok := value.([]interface{}); ok {
			log.Printf("ArrayItem Schema: %#v", self)
			for _, i := range base {
				e, files, err := self.Items.SchemaEvaluate(i)
				if err != nil {
					return []string{}, []JobFile{}, err
				}
				if self.Prefix != "" {
					out_args = append(out_args, self.Prefix)
				}
				out_args = append(out_args, e...)
				out_files = append(out_files, files...)
			}
		}
	} else {
		return []string{}, []JobFile{}, fmt.Errorf("Unknown Type '%s' (%#v)", typeName, *self)
	}
	if self.ItemSeparator != "" {
		out_args = []string{strings.Join(out_args, self.ItemSeparator)}
	}
	if self.Prefix != "" && typeName != "array" && typeName != "boolean" {
		out_args = append([]string{self.Prefix}, out_args...)
	}
	return out_args, out_files, nil
}

func (self *CommandInput) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	value_str := []string{}

	oFiles := []JobFile{}
	if base, ok := inputs[self.Id]; ok {
		var err error
		files := []JobFile{}
		value_str, files, err = self.SchemaEvaluate(base)
		if err != nil {
			log.Printf("Schema Evaluation Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		oFiles = append(oFiles, files...)
	} else {
		if self.Default != nil {
			var err error
			files := []JobFile{}
			value_str, files, err = self.SchemaEvaluate(*self.Default)
			if err != nil {
				log.Printf("Schema Evaluation Error: %s", err)
				return []string{}, []JobFile{}, err
			}
			oFiles = append(oFiles, files...)
		} else if self.IsOptional() {
			return []string{}, []JobFile{}, nil
		} else {
			return []string{}, []JobFile{}, fmt.Errorf("Input %s not found", self.Id)
		}
	}
	return value_str, oFiles, nil
}

func (self *CommandOutput) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	return []string{}, []JobFile{
		JobFile{
			Id:     self.Id,
			Output: true,
			Glob:   self.Glob,
		},
	}, nil
}

func (self *Argument) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	if self.Value != nil {
		return []string{*self.Value}, []JobFile{}, nil
	} else if self.ValueFrom != nil {
		//BUG: This is wrong. Need to evaluate expressions right before runtime (ie after paths are filed out)
		//But first need some structure to keep arrays togeather so they can be joined by ItemSeparator...
		exp, err := ExpressionEvaluate(*self.ValueFrom, inputs)
		if err != nil {
			log.Printf("Expression Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		value_str := []string{exp}
		if self.ItemSeparator != nil {
			value_str = []string{strings.Join(value_str, *self.ItemSeparator)}
		}
		if self.Prefix != nil {
			value_str = append([]string{*self.Prefix}, value_str...)
		}
		return value_str, []JobFile{}, nil
	}
	return []string{}, []JobFile{}, nil
}
