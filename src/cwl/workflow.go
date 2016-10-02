package cwl

import (
	"fmt"
	"log"
	"strings"
)

func (self Workflow) NewGraphState(inputs JSONDict) JSONDict {
	return JSONDict{INPUT_FIELD: inputs}
}

func (self Workflow) ReadySteps(state JSONDict) []string {
	out := []string{}
	for k, v := range self.Steps {
		if v.Ready(state) {
			log.Printf("Step Ready: %#v %#v", k, v.In)
			out = append(out, k)
		}
	}
	return out
}

func (self Workflow) UpdateStepResults(state JSONDict, stepId string, results JSONDict) JSONDict {
	out := JSONDict{}
	for k, v := range state {
		out[k] = v
	}
	step := JSONDict{}
	if base, ok := out[stepId].(JSONDict); ok {
		step = base
	} else {
		step[RESULTS_FIELD] = results
	}
	out[stepId] = step
	return out
}

func (self Workflow) Done(state JSONDict) bool {
	done := true
	for i := range self.Steps {
		if _, ok := state[i]; !ok {
			done = false
		}
	}
	log.Printf("CheckDone: %#v", done)
	return done
}

func (self Workflow) GenerateJob(step string, graphState JSONDict) (Job, error) {
	jobState := self.Steps[step].BuildStepInput(graphState)
	job, err := self.Steps[step].Doc.GenerateJob(step, jobState)
	if err != nil {
		return job, fmt.Errorf("Step %s failed: %s", step, err)
	}
	return job, err
}

func (self Workflow) GetIDs() []string {
	out := make([]string, 0, len(self.Steps))
	for _, k := range self.Steps {
		out = append(out, k.Id)
	}
	log.Printf("Workflow IDs: %#v", out)
	return out
}

func (self Workflow) GetResults(state JSONDict) JSONDict {
	log.Printf("Workflow Results: %#v", state)
	out := JSONDict{}
	for k, v := range self.Outputs {
		log.Printf("Workflow Output: %#v", v)
		out[k], _ = state.GetData(v.OutputSource)
	}
	return out
}

func (self Workflow) GetDefault(name string) (*interface{}, bool) {
	if strings.HasPrefix(name, "#") {
		name = name[:1]
	}
	if v, ok := self.Inputs[name]; ok {
		if v.Default != nil {
			return v.Default, true
		}
	}
	return nil, false
}

func (self Step) Ready(state JSONDict) bool {
	if _, ok := state.GetData(fmt.Sprintf("%s/%s", self.Id, RESULTS_FIELD)); ok {
		log.Printf("Step %s done", self.Id)
		return false
	}
	ready := true
	for _, v := range self.In {
		tmp := strings.Split(v.Source, "/")
		found := false
		if len(tmp) == 1 {
			if _, ok := state.GetData(fmt.Sprintf("%s/%s", INPUT_FIELD, tmp[0])); ok {
				found = true
			}
		} else {
			if b, ok := state[tmp[0]]; ok {
				o := self.Doc.GetResults(b.(JSONDict))
				log.Printf("Checking for %s in results %s", tmp[1], o)
				if _, ok := o[tmp[1]]; ok {
					found = true
				}
			}
		}
		if !found {
			if v.Default == nil {
				if _, ok := self.Parent.GetDefault(v.Source); !ok {
					ready = false
					log.Printf("Step %s input %s not found in %#v", self.Id, v.Source, state)
				} else {
					log.Printf("Step %s input %s has default", self.Id, v.Source)
				}
			}
		} else {
			log.Printf("Step %s found input %s in %#v", self.Id, v.Source, state)
		}
	}
	return ready
}

func (self Step) BuildStepInput(state JSONDict) JSONDict {
	out := JSONDict{}
	out[INPUT_FIELD] = JobState{}
	inputs := JSONDict{}
	for k, v := range self.In {
		tmp := strings.Split(v.Source, "/")
		if len(tmp) == 1 {
			if i, ok := state.GetData(fmt.Sprintf("%s/%s", INPUT_FIELD, tmp[0])); ok {
				inputs[k] = i
			}
		} else {
			if b, ok := state[tmp[0]]; ok {
				o := self.Doc.GetResults(b.(JSONDict))
				if i, ok := o[tmp[1]]; ok {
					inputs[k] = i
				}
			}
		}
		if _, ok := inputs[k]; !ok {
			if v.Default != nil {
				inputs[k] = *v.Default
			} else {
				i_default, _ := self.Parent.GetDefault(v.Source)
				if i_default != nil {
					inputs[k] = *i_default
				}
			}
		}
	}
	out[INPUT_FIELD] = inputs
	log.Printf("Input Built: %#v from %#v", out, self.In)
	return out
}
