package cwl

type CWLDocData map[string]interface{}

type JSONDict map[interface{}]interface{}

type GraphState map[string]JobState
type JobState map[string]interface{}

const INPUT_FIELD = "#"
const RESULTS_FIELD = "results"
const RUNTIME_FIELD = "runtime"
const ERROR_FIELD = "error"

type Job struct {
	Cmd    []string
	Files  []JobFile
	Stdout string
	Stderr string
	Stdin  string
	Inputs JSONDict
}

type JobFile struct {
	Id       string
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

type Workflow struct {
	Id      string
	Inputs  map[string]WorkflowInput
	Outputs map[string]WorkflowOutput
	Steps   map[string]Step
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
	Stdin        string
}

type cmdArg struct {
	position int
	id       string
	value    []string
}

type WorkflowInput struct {
	Id   string
	Type DataType
}

type WorkflowOutput struct {
	Id           string
	Type         DataType
	OutputSource string
}

type Step struct {
	Id  string
	In  map[string]string
	Out map[string]string
	Doc CWLDoc
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

type InlineJavascriptRequirement struct {
}

type InitialWorkDirRequirement struct {
}

type Argument struct {
	Value         *string
	ValueFrom     *string
	Position      int
	Prefix        *string
	ItemSeparator *string
}
