package cwl

import (
	"log"
)

func (self Workflow) NewGraphState(inputs JSONDict) GraphState {
	return GraphState{INPUT_FIELD: JobState{RESULTS_FIELD: inputs}}
}

func (self Workflow) ReadySteps(state GraphState) []string {
	out := []string{}
	for k, v := range self.Steps {
		if v.Ready(state) {
			log.Printf("Ready: %#v", k)
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
	return self.Steps[step].Doc.GenerateJob(step, graphState)
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
	return JSONDict{}
}

func (self Step) Ready(state GraphState) bool {
	if state.HasResults(self.Id) {
		return false
	}
	ready := true
	for _, v := range self.In {
		if !state.HasData(v) {
			ready = false
			log.Printf("Step %s input %s not found in %#v", self.Id, v, state)
		}
	}
	return ready
}

func (self Step) BuildStepInput(state GraphState) JSONDict {
	out := JSONDict{}
	for k, v := range self.In {
		out[k] = state.GetData(v)
	}
	return out
}
