package cwl

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"log"
	"regexp"
)

var EXP_RE_STRING, _ = regexp.Compile(`(.*)\$\((.*)\)(.*)`)
var EXP_RE, _ = regexp.Compile(`\$\((.*)\)`)

func (self *JSEvaluator) EvaluateExpressionString(expression string, js_self *JSONDict) (string, error) {

	matches := EXP_RE_STRING.FindStringSubmatch(expression)
	if matches == nil {
		return expression, nil
	}
	log.Printf("JS Expression: %s", matches[2])
	//log.Printf("JS Inputs: %#v", self.Inputs.Normalize())
	vm := otto.New()
	vm.Set("runtime", map[string]interface{}{"cores": 4})
	vm.Set("inputs", self.Inputs.Normalize())
	if js_self != nil {
		vm.Set("self", js_self.Normalize())
	}
	out, err := vm.Run(matches[2])
	log.Printf("JS:%s = %s\n", matches[2], out)
	return fmt.Sprintf("%s%s%s", matches[1], out, matches[3]), err
}

func otto2map(obj *otto.Object) JSONDict {
	out := JSONDict{}
	for _, i := range obj.Keys() {
		val, _ := obj.Get(i)
		if val.IsBoolean() {
			out[i], _ = val.ToBoolean()
		}
		if val.IsBoolean() {
			out[i], _ = val.ToBoolean()
		}
		if val.IsString() {
			out[i], _ = val.ToString()
		}
		if val.IsNumber() {
			out[i], _ = val.Export()
		}
		if val.IsObject() {
			out[i] = otto2map(val.Object())
		}
	}
	return out
}

func (self *JSEvaluator) EvaluateExpressionObject(expression string, js_self *JSONDict) (JSONDict, error) {

	matches := EXP_RE.FindStringSubmatch(expression)
	if matches == nil {
		return JSONDict{}, nil
	}
	log.Printf("JS Expression: %s", matches[1])
	vm := otto.New()
	vm.Set("runtime", map[string]interface{}{"cores": 4})
	ninputs := self.Inputs.Normalize()
	//log.Printf("Expression Inputs: %#v", ninputs)
	vm.Set("inputs", ninputs)
	if js_self != nil {
		vm.Set("self", js_self.Normalize())
	}
	out, err := vm.Run("out=" + matches[1])
	return otto2map(out.Object()), err
}
