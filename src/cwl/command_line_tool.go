package cwl

import (
	"log"
	"sort"
	"fmt"
	"strings"
)

func (self CommandLineTool) NewGraphState(inputs JSONDict) GraphState {
	return GraphState{"#":inputs}
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
		cmd, err := self.Evaluate(i.(JSONDict))
		if err != nil {
			log.Printf("Job Eval Error: %s", err)
			return Job{}, err
		}
		return Job{Cmd:cmd}, nil
	}
}


func (self *CommandLineTool) Evaluate(inputs JSONDict) ([]string, error) {
	log.Printf("CommandLineTool Evalute")
	out := make([]string, 0)
	out = append(out, self.BaseCommand...)

	args := make(cmdArgArray, 0, len(self.Arguments)+len(self.Inputs))
	//Arguments
	for _, x := range self.Arguments {
		e, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Argument Error: %s", err)
			return []string{}, err
		}
		c := cmdArg{
			position: x.Position,
			value:    e,
		}
		args = append(args, c)
	}
	//Inputs
	for _, x := range self.Inputs {
		e, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Input Error: %s", err)
			return []string{}, err
		}
		c := cmdArg{
			position: x.Position,
			value:    e,
		}
		args = append(args, c)
	}

	sort.Stable(args)
	for _, x := range args {
		out = append(out, x.value...)
	}
	log.Printf("Out: %v", out)
	return out, nil
}


func (self *DataType) Evaluate(value interface{}) ([]string, error) {
	if self.TypeName == "File" {
		if base, ok := value.(map[interface{}]interface{}); ok {
			if class, ok := base["class"]; ok {
				if class.(string) == "File" {
					return []string{base["location"].(string)}, nil
				}
			}
		}
	} else if self.TypeName == "int" {
		return []string{fmt.Sprintf("%d", value.(int))}, nil
	} else if self.TypeName == "array" {
		out := []string{}
		if base, ok := value.([]interface{}); ok {
			for _, i := range base {
				e, err := self.Items.Evaluate(i)
				if err != nil {
					return []string{}, err
				}
				out = append(out, e...)
			}
		}
		return out, nil
	}
	return []string{}, fmt.Errorf("Unknown Type %s", self.TypeName)
}

func (self *CommandInput) Evaluate(inputs JSONDict) ([]string, error) {
	value_str := []string{}

	if base, ok := inputs[self.Id]; ok {
		var err error
		value_str, err = self.Type.Evaluate(base)
		if err != nil {
			log.Printf("DataType Evaluation Error: %s", err)
			return []string{}, err
		}
	} else {
		if self.Default != nil {
			var err error
			value_str, err = self.Type.Evaluate(*self.Default)
			if err != nil {
				log.Printf("DataType Evaluation Error: %s", err)
				return []string{}, err
			}			
		} else {
			return []string{}, fmt.Errorf("Input %s not found", self.Id)
		}
	}
	
	if self.ItemSeparator != nil {
		value_str = []string{strings.Join(value_str, *self.ItemSeparator)}
	}
	if self.Prefix != nil {
		value_str = append([]string{*self.Prefix}, value_str...)
	}
	return value_str, nil

}


func (self *Argument) Evaluate(inputs JSONDict) ([]string, error) {
	//fmt.Printf("Arguments %#v\n", self)
	if self.Value != nil {
		return []string{*self.Value}, nil
	} else if self.ValueFrom != nil {
		//fmt.Println("Arguments Evaluate")
		exp, err := ExpressionEvaluate(*self.ValueFrom, inputs)
		if err != nil {
			log.Printf("Expression Error: %s", err)
			return []string{}, err
		}
		value_str := []string{exp}
		if self.ItemSeparator != nil {
			value_str = []string{strings.Join(value_str, *self.ItemSeparator)}
		}
		if self.Prefix != nil {
			value_str = append([]string{*self.Prefix}, value_str...)
		}
		return value_str, nil
	}
	return []string{}, nil
}

