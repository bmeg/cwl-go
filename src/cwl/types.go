package cwl

type CWLDocData map[string]interface{}

type JSONDict map[interface{}]interface{}

type GraphState map[string]JobState
type JobState map[string]interface{}

const INPUT_FIELD = "#"
const RESULTS_FIELD = "results"
const RUNTIME_FIELD = "runtime"
const ERROR_FIELD = "error"

const (
	COMMAND    = iota
	EXPRESSION = iota
)

type Job struct {
	JobType      int
	Cmd          []JobArgument
	Expression   string
	Stdout       string
	Stderr       string
	Stdin        string
	InputData    JSONDict
	Inputs       map[string]Schema
	Outputs      map[string]Schema
	SuccessCodes []int
}

type JSEvaluator struct {
	Inputs  JSONDict
	Runtime JSONDict
	Outputs JSONDict
}

type JobArgument struct {
	Id       string
	Position int
	Value    string
	RawValue interface{}
	Join     string
	Prefix   string
	File     *JobFile
	Children []JobArgument
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
	SuccessCodes []int
}

type WorkflowInput struct {
	Schema
	Source string
}

type WorkflowOutput struct {
	Schema
	OutputSource string
}

type Step struct {
	Id     string
	In     map[string]StepInput
	Out    map[string]StepOutput
	Doc    CWLDoc
	Parent *Workflow
}

type StepInput struct {
	Id      string
	Source  string
	Default *interface{}
}

type StepOutput struct {
	Id string
}

type Schema struct {
	Id            string
	Name          string
	TypeName      string
	Items         *Schema
	Types         []Schema
	Prefix        string
	Position      int
	ItemSeparator string
	Default       *interface{}
}

type CommandInput struct {
	Schema
}

type CommandOutput struct {
	Schema
	Glob string
}

type ExpressionTool struct {
	Id           string
	Inputs       map[string]ExpressionInput
	Outputs      map[string]ExpressionOutput
	Expression   string
	Requirements []Requirement
}

type ExpressionInput struct {
	Schema
}

type ExpressionOutput struct {
	Schema
}

type Requirement interface{}

type UnsupportedRequirement struct {
	Message string
}

func (e UnsupportedRequirement) Error() string {
	return e.Message
}

type SchemaDefRequirement struct {
	NewTypes []Schema
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
