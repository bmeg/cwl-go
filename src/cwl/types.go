package cwl

type CWLDocData map[string]interface{}

type JSONDict map[interface{}]interface{}

type GraphState map[string]interface{}

type Job struct {
	Cmd    []string
	Files  []JobFile
	Stdout string
	Stderr string
}

type JobFile struct {
	Path     string
	Location string
	Dir      bool
	Output   bool
	Glob     string
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
	Stdout       string
	Stderr       string
}

type cmdArg struct {
	position int
	id       string
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
	Prefix   *string
}

type CommandOutput struct {
	Id   string
	Type DataType
	Glob string
}

type Requirement interface{}

type SchemaDefRequirement struct {
	NewTypes []DataType
}

type Argument struct {
	Value         *string
	ValueFrom     *string
	Position      int
	Prefix        *string
	ItemSeparator *string
}
