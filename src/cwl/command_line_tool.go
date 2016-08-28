package cwl

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

func (self CommandLineTool) NewGraphState(inputs JSONDict) GraphState {
	return GraphState{"#": inputs}
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
	out[stepId] = results
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
	return state[self.Id].(JSONDict)
}

func (self CommandLineTool) GenerateJob(step string, graphState GraphState) (Job, error) {
	if i, ok := graphState["#"]; !ok {
		return Job{}, fmt.Errorf("%s Inputs not ready", step)
	} else {
		cmd, files, err := self.Evaluate(i.(JSONDict))
		if err != nil {
			log.Printf("Job Eval Error: %s", err)
			return Job{}, err
		}
		return Job{Cmd: cmd, Files: files}, nil
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

	sort.Stable(args)
	for _, x := range args {
		out = append(out, x.value...)
	}
	log.Printf("Out: %v", out)
	return out, oFiles, nil
}

func (self *DataType) Evaluate(value interface{}) ([]string, []JobFile, error) {
	if self.TypeName == "File" {
		if base, ok := value.(map[interface{}]interface{}); ok {
			if class, ok := base["class"]; ok {
				if class.(string) == "File" {
					loc := base["location"].(string)
					return []string{loc}, []JobFile{JobFile{Location: loc}}, nil
				} else {
					log.Printf("Unknown class %s", class)
				}
			} else {
				log.Printf("Input map has no class")
			}
		} else {
			log.Printf("File input not formatted correctly: %#v", value)
		}
	} else if self.TypeName == "int" {
		return []string{fmt.Sprintf("%d", value.(int))}, []JobFile{}, nil
	} else if self.TypeName == "array" {
		out := []string{}
		oFiles := []JobFile{}
		if base, ok := value.([]interface{}); ok {
			for _, i := range base {
				e, files, err := self.Items.Evaluate(i)
				if err != nil {
					return []string{}, []JobFile{}, err
				}
				if self.Prefix != nil {
					out = append(out, *self.Prefix)
				}
				out = append(out, e...)
				oFiles = append(oFiles, files...)
			}
		}
		return out, oFiles, nil
	}
	return []string{}, []JobFile{}, fmt.Errorf("Unknown Type %s", self.TypeName)
}

func (self *CommandInput) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	value_str := []string{}

	oFiles := []JobFile{}
	if base, ok := inputs[self.Id]; ok {
		var err error
		files := []JobFile{}
		value_str, files, err = self.Type.Evaluate(base)
		if err != nil {
			log.Printf("DataType Evaluation Error: %s", err)
			return []string{}, []JobFile{}, err
		}
		oFiles = append(oFiles, files...)
	} else {
		if self.Default != nil {
			var err error
			files := []JobFile{}
			value_str, files, err = self.Type.Evaluate(*self.Default)
			if err != nil {
				log.Printf("DataType Evaluation Error: %s", err)
				return []string{}, []JobFile{}, err
			}
			oFiles = append(oFiles, files...)
		} else {
			return []string{}, []JobFile{}, fmt.Errorf("Input %s not found", self.Id)
		}
	}

	if self.ItemSeparator != nil {
		value_str = []string{strings.Join(value_str, *self.ItemSeparator)}
	}
	if self.Prefix != nil {
		value_str = append([]string{*self.Prefix}, value_str...)
	}
	return value_str, oFiles, nil

}

func (self *Argument) Evaluate(inputs JSONDict) ([]string, []JobFile, error) {
	//fmt.Printf("Arguments %#v\n", self)
	if self.Value != nil {
		return []string{*self.Value}, []JobFile{}, nil
	} else if self.ValueFrom != nil {
		//fmt.Println("Arguments Evaluate")
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
