package cwl

type CWLDocData map[string]interface{}

type JSONDict map[string]interface{}

type GraphState map[string]interface{}

type Job struct {
	Cmd   []string
	Files []JobFile
}

type JobFile struct {
	Path     string
	Location string
	Dir      bool
	Output   bool
}

type CWLDoc interface {
	GetIDs() []string
	NewGraphState(inputs JSONDict) GraphState
	Done(GraphState) bool
	UpdateStepResults(GraphState, string, JSONDict) GraphState
	ReadySteps(state GraphState) []string
	GetResults(state GraphState) JSONDict
	GenerateJob(step string, graphState GraphState) (Job, error)
}

type CommandLineTool struct {
	Id           string
	Inputs       map[string]CommandInput
	Outputs      map[string]CommandOutput
	BaseCommand  []string
	Requirements []Requirement
	Arguments    []Argument
}

type cmdArg struct {
	position int
	value    []string
}

type CommandInput struct {
	Id            string
	Position      int
	Prefix        *string
	ItemSeparator *string
	Type          DataType
	Default       *interface{}
}

type DataType struct {
	TypeName string
	Items    *DataType
}

type CommandOutput struct{}

type Requirement struct{}

type Argument struct {
	Value         *string
	ValueFrom     *string
	Position      int
	Prefix        *string
	ItemSeparator *string
}
