package cwl

import (
	"log"
	"strings"
)

func (self *Job) GetFiles() []JobFile {
	out := []JobFile{}
	for _, e := range self.Cmd {
		out = append(out, e.GetFiles()...)
	}
	return out
}

func (self *JobArgument) GetFiles() []JobFile {
	out := []JobFile{}
	if self.File != nil {
		out = append(out, *self.File)
	}
	for _, x := range self.Children {
		out = append(out, x.GetFiles()...)
	}
	return out
}

func (self *JobArgument) GetArgs(evaluator JSEvaluator, pathMapper func(interface{}) interface{}) ([]string, error) {
	cmd, bound, err := self.EvaluateStrings(evaluator, pathMapper)
	if err != nil || !bound {
		return []string{}, err
	}
	return cmd, nil
}

func (self *JobArgument) EvaluateStrings(evaluator JSEvaluator, pathMapper func(interface{}) interface{}) ([]string, bool, error) {
	out := []string{}

	if self.Value != "" {
		var e string
		var err error
		if self.File != nil {
			f := pathMapper(self.File.ToJSONDict()).(JSONDict)
			log.Printf("self = %s", f)
			e, err = evaluator.EvaluateExpressionString(self.Value, &f)
		} else {
			e, err = evaluator.EvaluateExpressionString(self.Value, nil)
		}
		if err != nil {
			return []string{}, false, err
		}
		out = append(out, e)
	} else if self.Bound && self.File != nil {
		out = append(out, self.File.Location)
	}
	for _, x := range self.Children {
		e, _, err := x.EvaluateStrings(evaluator, pathMapper)
		if err != nil {
			return []string{}, false, err
		}
		out = append(out, e...)
	}

	if self.Join != "" {
		out = []string{strings.Join(out, self.Join)}
	}

	if self.Prefix != "" {
		out = append([]string{self.Prefix}, out...)
	}

	return out, self.Bound, nil
}

func (self *JobArgument) EvaluateObject(evaluator JSEvaluator) (interface{}, error) {
	log.Printf("Expression:%s", self.RawValue)
	return self.RawValue, nil
}
