package cwl

import (
	"fmt"
	"log"
)

func (self ExpressionTool) NewGraphState(inputs JSONDict) JSONDict {
	return JSONDict{INPUT_FIELD: inputs}
}

func (self ExpressionTool) Done(state JSONDict) bool {
	if _, ok := state[RESULTS_FIELD]; ok {
		return true
	}
	return false
}

func (self ExpressionTool) GenerateJob(step string, graphState JSONDict) (Job, error) {
	outputs := map[string]Schema{}
	for k, v := range self.Outputs {
		outputs[k] = v.Schema
	}
	inputs := map[string]Schema{}
	for k, v := range self.Inputs {
		inputs[k] = v.Schema
	}

	args := []JobArgument{}
	if i, ok := graphState[INPUT_FIELD]; !ok {
		return Job{}, fmt.Errorf("%s Inputs not ready", step)
	} else {
		data_input := i.(JSONDict)
		for _, x := range self.Inputs {
			new_args, err := x.Evaluate(data_input)
			if err != nil {
				return Job{}, err
			}
			args = append(args, new_args)
		}
	}

	return Job{
		JobType:    EXPRESSION,
		Cmd:        args,
		Expression: self.Expression,
		Outputs:    outputs,
		Inputs:     inputs,
		InputData:  graphState[INPUT_FIELD].(JSONDict),
	}, nil
}

func (self ExpressionTool) GetIDs() []string {
	return []string{self.Id}
}

func (self ExpressionTool) GetResults(state JSONDict) JSONDict {
	return state[RESULTS_FIELD].(JSONDict)
}

func (self ExpressionTool) ReadySteps(state JSONDict) []string {
	if _, ok := state[INPUT_FIELD]; ok {
		return []string{self.Id}
	}
	return []string{}
}

func (self ExpressionTool) UpdateStepResults(state JSONDict, stepId string, results JSONDict) JSONDict {
	out := JSONDict{}
	for k, v := range state {
		out[k] = v
	}
	out[RESULTS_FIELD] = results
	return out
}

func (self *ExpressionInput) Evaluate(inputs JSONDict) (JobArgument, error) {
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
