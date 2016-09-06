package cwl

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"log"
	"regexp"
)

var EXP_RE, _ = regexp.Compile(`\$\((.*)\)`)

func (self *JSEvaluator) EvaluateExpression(expression string, js_self *JSONDict) (string, error) {

	matches := EXP_RE.FindStringSubmatch(expression)
	if matches == nil {
		return expression, nil
	}
	log.Printf("JS Expression: %s", matches[1])
	vm := otto.New()
	vm.Set("runtime", map[string]interface{}{"cores": 4})
	vm.Set("inputs", self.Inputs.Normalize())
	if js_self != nil {
		vm.Set("self", js_self.Normalize())
	}
	out, err := vm.Run(matches[1])
	log.Printf("JS:%s = %s\n", matches[1], out)
	return fmt.Sprintf("%s", out), err
}
