package cwl

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"log"
	"regexp"
)

func ExpressionEvaluate(expression string, inputs JSONDict) (string, error) {
	exp_re, _ := regexp.Compile(`\$\((.*)\)`)
	matches := exp_re.FindStringSubmatch(expression)
	log.Printf("Matches: %s %#v", expression, matches)
	vm := otto.New()
	vm.Set("runtime", map[string]interface{}{"cores": 4})
	out, err := vm.Run(matches[1])
	log.Printf("JS:%s = %s\n", matches[1], out)
	return fmt.Sprintf("%s", out), err
}
