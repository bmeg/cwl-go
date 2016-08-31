package cwl

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"log"
	"regexp"
)

var EXP_RE, _ = regexp.Compile(`\$\((.*)\)`)

func ExpressionEvaluate(expression string, inputs JSONDict) (string, error) {

	matches := EXP_RE.FindStringSubmatch(expression)
	if matches == nil {
		return expression, nil
	}
	log.Printf("JS Expression: %s", matches[1])
	vm := otto.New()
	vm.Set("runtime", map[string]interface{}{"cores": 4})
	vm.Set("inputs", inputs.Normalize())
	out, err := vm.Run(matches[1])
	log.Printf("JS:%s = %s\n", matches[1], out)
	return fmt.Sprintf("%s", out), err
}
