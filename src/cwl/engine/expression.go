package cwl_engine

import (
	"cwl"
	"fmt"
	"log"
)

func NewExpressionRunner(config Config) JobRunner {
	return ExpressionRunner{Config: config}
}

type ExpressionRunner struct {
	Config Config
}

func (self ExpressionRunner) GetWorkDirPath() string {
	return "/tmp"
}
func (self ExpressionRunner) LocationToPath(location string) string {
	return location
}

func (self ExpressionRunner) Glob(string) []string {
	return []string{}
}

func (self ExpressionRunner) ReadFile(string) ([]byte, error) {
	return []byte{}, fmt.Errorf("No files in expression engine")
}

func (self ExpressionRunner) StartProcess(inputs cwl.JSONDict, cmd_args []string, workdir, stdout, stderr, stdin, dockerImage string) (cwl.JSONDict, error) {
	log.Printf("Running Expression %s", cmd_args[0])
	log.Printf("Expression Inputs: %#v", inputs)

	js_eval := cwl.JSEvaluator{Inputs: inputs}

	out, err := js_eval.EvaluateExpressionObject(cmd_args[0], nil)
	if err != nil {
		return cwl.JSONDict{}, fmt.Errorf("ExpressionTool Failure: %s", err)
	}
	log.Printf("expression out: %s", out)
	return cwl.JSONDict{"output": out}, nil
}

func (self ExpressionRunner) ExitCode(prodData cwl.JSONDict) (int, bool) {
	return 0, true
}

func (self ExpressionRunner) GetOutput(prodData cwl.JSONDict) cwl.JSONDict {

	return prodData["output"].(cwl.JSONDict)
}
