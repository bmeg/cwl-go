package cwl

import (
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

func (self *JobArgument) Evaluate(evaluator JSEvaluator) ([]string, error) {
	out := []string{}

	if self.Value != "" {
		var e string
		var err error
		if self.File != nil {
			f := self.File.ToJSONDict()
			e, err = evaluator.EvaluateExpression(self.Value, &f)
		} else {
			e, err = evaluator.EvaluateExpression(self.Value, nil)
		}
		if err != nil {
			return []string{}, err
		}
		out = append(out, e)
	}

	for _, x := range self.Children {
		e, err := x.Evaluate(evaluator)
		if err != nil {
			return []string{}, err
		}
		out = append(out, e...)
	}

	if self.Join != "" {
		out = []string{strings.Join(out, self.Join)}
	}

	if self.Prefix != "" {
		out = append([]string{self.Prefix}, out...)
	}

	return out, nil
}
