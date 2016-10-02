package cwl

type CWLDocData map[string]interface{}

type JSONDict map[interface{}]interface{}

type JobState map[string]interface{}

const INPUT_FIELD = "#"
const RESULTS_FIELD = "?"
const RUNTIME_FIELD = "@"
const ERROR_FIELD = "!"

const (
	COMMAND    = iota
	EXPRESSION = iota
)

type Job struct {
	JobType      int
	Cmd          []JobArgument
	DockerImage  string
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
	Bound    bool
	File     *JobFile
	Children []JobArgument
}

type JobFile struct {
	Id           string
	Path         string
	Location     string
	Dir          bool
	Output       bool
	LoadContents bool
	Glob         string
}

type CWLGraph struct {
	Elements map[string]CWLDoc
	Main     string
}

type CWLDoc interface {
	GetIDs() []string
	NewGraphState(inputs JSONDict) JSONDict
	Done(JSONDict) bool
	UpdateStepResults(JSONDict, string, JSONDict) JSONDict
	ReadySteps(state JSONDict) []string
	GetResults(state JSONDict) JSONDict
	GenerateJob(step string, graphState JSONDict) (Job, error)
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
	Schema
	Source string
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
	Bound         bool
	LoadContents  bool
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

type DockerRequirement struct {
	DockerPull string
}

type ResourceRequirement struct {
	Props map[string]interface{}
}

type InlineJavascriptRequirement struct {
}

type InitialWorkDirRequirement struct {
}

type Argument struct {
	Schema
	Value     *string
	ValueFrom *string
}
