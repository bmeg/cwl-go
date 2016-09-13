package cwl

import (
	"fmt"
	"log"
)

func (self Workflow) NewGraphState(inputs JSONDict) GraphState {
	return GraphState{INPUT_FIELD: JobState{RESULTS_FIELD: inputs}}
}

func (self Workflow) ReadySteps(state GraphState) []string {
	out := []string{}
	for k, v := range self.Steps {
		if v.Ready(state) {
			log.Printf("Step Ready: %#v %#v", k, v.In)
			out = append(out, k)
		}
	}
	return out
}

func (self Workflow) UpdateStepResults(state GraphState, stepId string, results JSONDict) GraphState {
	out := GraphState{}
	for k, v := range state {
		out[k] = v
	}
	out[stepId] = JobState{RESULTS_FIELD: results}
	return out
}

func (self Workflow) Done(state GraphState) bool {
	done := true
	for i := range self.Steps {
		if _, ok := state[i]; !ok {
			done = false
		}
	}
	log.Printf("CheckDone: %#v", done)
	return done
}

func (self Workflow) GenerateJob(step string, graphState GraphState) (Job, error) {
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

func (self Workflow) GetResults(state GraphState) JSONDict {
	out := JSONDict{}
	for k, v := range self.Outputs {
		log.Printf("Workflow Output: %#v", v)
		out[k], _ = state.GetData(v.OutputSource)
	}
	return out
}

func (self Step) Ready(state GraphState) bool {
	if state.HasResults(self.Id) {
		return false
	}
	ready := true
	for _, v := range self.In {
		if _, ok := state.GetData(v.Source); !ok {
			ready = false
			log.Printf("Step %s input %s not found in %#v", self.Id, v.Source, state)
		} else {
			log.Printf("Step %s found input %s in %#v", self.Id, v.Source, state)
		}
	}
	return ready
}

func (self Step) BuildStepInput(state GraphState) GraphState {
	out := GraphState{}
	out[INPUT_FIELD] = JobState{}
	inputs := JSONDict{}
	for k, v := range self.In {
		inputs[k], _ = state.GetData(v.Source)
	}
	out[INPUT_FIELD][RESULTS_FIELD] = inputs
	log.Printf("Input Built: %#v from %#v", out, self.In)
	return out
}
