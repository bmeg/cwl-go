package cwl

import (
	"fmt"
	"log"
	"sort"
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
		args, err := self.Evaluate(i[RESULTS_FIELD].(JSONDict))
		if err != nil {
			log.Printf("Job Eval Error: %s", err)
			return Job{}, err
		}
		stdout := self.Stdout
		stderr := self.Stderr
		stdin := self.Stdin

		outputs := map[string]Schema{}
		for k, v := range self.Outputs {
			outputs[k] = v.Schema
			//for stdout/stderr shortcuts (that have the type stdout/stderr)
			//create the temp files that will be used to store the output
			if v.Schema.TypeName == "stdout" {
				if stdout == "" {
					stdout = "__cwl_stdout"
				}
				for i := range args {
					if args[i].Id == outputs[k].Id {
						args[i].File.Glob = stdout
					}
				}
			}
			if v.Schema.TypeName == "stderr" {
				if stderr == "" {
					stderr = "__cwl_stderr"
				}
				for i := range args {
					if args[i].Id == outputs[k].Id {
						args[i].File.Glob = stderr
					}
				}
			}
		}
		inputs := map[string]Schema{}
		for k, v := range self.Inputs {
			inputs[k] = v.Schema
		}

		dockerImage := ""
		for _, i := range self.Requirements {
			if a, ok := i.(DockerRequirement); ok {
				dockerImage = a.DockerPull
			}
		}

		return Job{JobType: COMMAND,
			Cmd:          args,
			Stderr:       stderr,
			Stdout:       stdout,
			Stdin:        stdin,
			InputData:    i[RESULTS_FIELD].(JSONDict),
			DockerImage:  dockerImage,
			SuccessCodes: self.SuccessCodes,
			Outputs:      outputs,
			Inputs:       inputs,
		}, nil
	}
}

type jobArgArray []JobArgument

func (self jobArgArray) Len() int {
	return len(self)
}

func (self jobArgArray) Less(i, j int) bool {
	if (self)[i].Position == (self)[j].Position {
		return (self)[i].Id < (self)[j].Id
	}
	return (self)[i].Position < (self)[j].Position
}

func (self jobArgArray) Swap(i, j int) {
	(self)[i], (self)[j] = (self)[j], (self)[i]
}

func (self *CommandLineTool) Evaluate(inputs JSONDict) ([]JobArgument, error) {
	log.Printf("CommandLineTool Evalute")

	args := make(jobArgArray, 0, len(self.Arguments)+len(self.Inputs))

	for _, x := range self.BaseCommand {
		args = append(args, JobArgument{Value: x, Position: -10000, Bound: true})
	}

	//Arguments
	for _, x := range self.Arguments {
		new_args, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Argument Error: %s", err)
			return []JobArgument{}, err
		}
		args = append(args, new_args)
	}
	//Inputs
	for _, x := range self.Inputs {
		new_args, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Input Error: %s", err)
			return []JobArgument{}, err
		}
		args = append(args, new_args)
	}

	//Outputs
	for _, x := range self.Outputs {
		new_args, err := x.Evaluate(inputs)
		if err != nil {
			log.Printf("Output Error: %s", err)
			return []JobArgument{}, err
		}
		args = append(args, new_args)
	}

	sort.Stable(args)
	//log.Printf("Out: %v", args)
	return args, nil
}

func (self *Schema) IsOptional() bool {
	for _, a := range self.Types {
		if a.TypeName == "null" {
			return true
		}
	}
	return false
}

func (self *Schema) SchemaEvaluate(value interface{}) (JobArgument, error) {
	out_args := JobArgument{
		Id:       self.Id,
		Join:     self.ItemSeparator,
		Position: self.Position,
		Bound:    self.Bound,
		RawValue: value,
	}

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
					out_args.File = &JobFile{Id: self.Id, Location: loc}
					out_args.Value = "$(self.path)"
				} else {
					log.Printf("Unknown class %s", class)
				}
			} else {
				log.Printf("Input map has no class")
			}
		} else {
			log.Printf("File input not formatted correctly: %#v", value)
		}
	} else if typeName == "Directory" {
		if base, ok := value.(map[interface{}]interface{}); ok {
			if class, ok := base["class"]; ok {
				if class.(string) == "Directory" {
					loc := base["location"].(string)
					out_args.File = &JobFile{Id: self.Id, Location: loc, Dir: true}
					out_args.Value = "$(self.path)"
				} else {
					log.Printf("Unknown class %s", class)
				}
			} else {
				log.Printf("Input map has no class: %#v", base)
			}
		} else {
			log.Printf("Directory input not formatted correctly: %#v", value)
		}
	} else if typeName == "int" {
		out_args.Value = fmt.Sprintf("%d", value.(int))
	} else if typeName == "boolean" {
		if value.(bool) {
			out_args.Prefix = self.Prefix
		} else {
			out_args.Prefix = ""
		}
	} else if typeName == "Any" {
		out_args.Value = fmt.Sprintf("%s", value)
	} else if typeName == "array_holder" {
		o, err := self.Types[0].SchemaEvaluate(value)
		if err != nil {
			return JobArgument{}, fmt.Errorf("Bad array '%s' (%#v): %s", typeName, *self, err)
		}
		out_args = o
		out_args.Join = self.ItemSeparator
		out_args.Position = self.Position
		out_args.Bound = self.Bound
	} else if typeName == "array" {
		if base, ok := value.([]interface{}); ok {
			log.Printf("Evalutate ArrayItem Schema: %#v", self)
			for _, i := range base {
				e, err := self.Items.SchemaEvaluate(i)
				if err != nil {
					return JobArgument{}, err
				}
				if self.Prefix != "" {
					out_args.Children = append(out_args.Children, JobArgument{Id: self.Id, Value: self.Prefix})
				}
				out_args.Children = append(out_args.Children, e)
			}
		}
	} else {
		return JobArgument{}, fmt.Errorf("Unknown Type '%s' (%#v)", typeName, *self)
	}
	if self.Prefix != "" && typeName != "array" && typeName != "boolean" {
		out_args.Prefix = self.Prefix
	}
	return out_args, nil
}

func (self *CommandInput) Evaluate(inputs JSONDict) (JobArgument, error) {
	out_arg := JobArgument{}

	if base, ok := inputs[self.Id]; ok {
		a, err := self.SchemaEvaluate(base)
		if err != nil {
			log.Printf("Schema Evaluation Error: %s", err)
			return JobArgument{}, err
		}
		out_arg = a
	} else {
		if self.Default != nil {
			a, err := self.SchemaEvaluate(*self.Default)
			if err != nil {
				log.Printf("Schema Evaluation Error: %s", err)
				return JobArgument{}, err
			}
			out_arg = a
			log.Printf("Default Eval: %s %#v", self.Id, out_arg)
		} else if self.IsOptional() {
			return JobArgument{}, nil
		} else {
			return JobArgument{}, fmt.Errorf("Input '%s' not found in %#v", self.Id, inputs)
		}
	}
	return out_arg, nil
}

func (self *CommandOutput) Evaluate(inputs JSONDict) (JobArgument, error) {
	return JobArgument{
		Id: self.Id,
		File: &JobFile{
			Id:     self.Id,
			Output: true,
			Glob:   self.Glob,
		},
	}, nil
}

func (self *Argument) Evaluate(inputs JSONDict) (JobArgument, error) {
	if self.Value != nil {
		return JobArgument{Value: *self.Value, Bound: self.Bound}, nil
	} else if self.ValueFrom != nil {
		out := JobArgument{Value: *self.ValueFrom, Bound: self.Bound}
		if self.Prefix != "" {
			out.Prefix = self.Prefix
		}
		out.Position = self.Position
		return out, nil
	}
	return JobArgument{}, nil
}
