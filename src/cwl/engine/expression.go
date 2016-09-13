package cwl_engine

import (
	"cwl"
	"fmt"
	"log"
)

func NewExpressionRunner(config Config) CWLRunner {
	return ExpressionRunner{Config: config}
}

type ExpressionRunner struct {
	Config Config
}

func (self ExpressionRunner) LocationToPath(location string) string {
	return location
}

func (self ExpressionRunner) RunCommand(job cwl.Job) (cwl.JSONDict, error) {
	log.Printf("Running Expression %s", job.Expression)

	inputs := MapInputs(job.InputData, self)
	js_eval := cwl.JSEvaluator{Inputs: inputs}

	js_inputs := cwl.JSONDict{}
	for _, v := range job.Cmd {
		s, err := v.EvaluateObject(js_eval)
		if err != nil {
			return cwl.JSONDict{}, fmt.Errorf("Input Eval Failure: %s", err)
		}
		js_inputs[v.Id] = s
	}
	js_eval = cwl.JSEvaluator{Inputs: js_inputs}

	log.Printf("Expression Inputs: %#v", js_inputs)
	out, err := js_eval.EvaluateExpressionObject(job.Expression, nil)
	if err != nil {
		return cwl.JSONDict{}, fmt.Errorf("ExpressionTool Failure: %s", err)
	}
	log.Printf("expression out: %s", out)

	out_dict := cwl.JSONDict{}
	for k, v := range job.Outputs {
		o_v, _ := out.Get(k)
		log.Printf("%s: %#v", k, v, o_v)
		switch {
		case v.TypeName == "int":
			var err error
			out_dict[k], err = o_v.ToInteger()
			if err != nil {
				return cwl.JSONDict{}, fmt.Errorf("Output Read error Type: %s", err)
			}
		default:
			return cwl.JSONDict{}, fmt.Errorf("Unknown Type: %s", v.TypeName)
		}
	}
	log.Printf("Expression Returning: %#v", out_dict)
	return out_dict, nil
}
