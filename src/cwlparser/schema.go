
package cwlparser

import (
  "reflect"
)


type CWLVersion string
const (
    CWLVersion_V1_0 CWLVersion = "v1.0"
    CWLVersion_Draft_3_dev1 CWLVersion = "draft-3.dev1"
    CWLVersion_Draft_3_dev3 CWLVersion = "draft-3.dev3"
    CWLVersion_Draft_3_dev2 CWLVersion = "draft-3.dev2"
    CWLVersion_Draft_3_dev5 CWLVersion = "draft-3.dev5"
    CWLVersion_Draft_3_dev4 CWLVersion = "draft-3.dev4"
    CWLVersion_Draft_3 CWLVersion = "draft-3"
    CWLVersion_Draft_2 CWLVersion = "draft-2"
    CWLVersion_V1_0_dev4 CWLVersion = "v1.0.dev4"
    CWLVersion_Draft_4_dev1 CWLVersion = "draft-4.dev1"
    CWLVersion_Draft_4_dev2 CWLVersion = "draft-4.dev2"
    CWLVersion_Draft_4_dev3 CWLVersion = "draft-4.dev3"
)

type CWLType string
const (
    CWLType_Directory CWLType = "Directory"
    CWLType_File CWLType = "File"
)

type LinkMergeMethod string
const (
    LinkMergeMethod_Merge_flattened LinkMergeMethod = "merge_flattened"
    LinkMergeMethod_Merge_nested LinkMergeMethod = "merge_nested"
)

type ScatterMethod string
const (
    ScatterMethod_Dotproduct ScatterMethod = "dotproduct"
    ScatterMethod_Nested_crossproduct ScatterMethod = "nested_crossproduct"
    ScatterMethod_Flat_crossproduct ScatterMethod = "flat_crossproduct"
)

type PrimitiveType string
const (
    PrimitiveType_String PrimitiveType = "string"
    PrimitiveType_Int PrimitiveType = "int"
    PrimitiveType_Double PrimitiveType = "double"
    PrimitiveType_Float PrimitiveType = "float"
    PrimitiveType_Long PrimitiveType = "long"
    PrimitiveType_Boolean PrimitiveType = "boolean"
    PrimitiveType_Null PrimitiveType = "null"
)

type Expression string
const (
    Expression_ExpressionPlaceholder Expression = "ExpressionPlaceholder"
)

var TYPES = map[string]reflect.Type {
    "CommandInputRecordField" : reflect.TypeOf(CommandInputRecordField{}),
    "OutputEnumSchema" : reflect.TypeOf(OutputEnumSchema{}),
    "CommandInputArraySchema" : reflect.TypeOf(CommandInputArraySchema{}),
    "EnumSchema" : reflect.TypeOf(EnumSchema{}),
    "ExpressionToolOutputParameter" : reflect.TypeOf(ExpressionToolOutputParameter{}),
    "WorkflowStepInput" : reflect.TypeOf(WorkflowStepInput{}),
    "InputRecordSchema" : reflect.TypeOf(InputRecordSchema{}),
    "WorkflowStepOutput" : reflect.TypeOf(WorkflowStepOutput{}),
    "OutputArraySchema" : reflect.TypeOf(OutputArraySchema{}),
    "CommandLineBinding" : reflect.TypeOf(CommandLineBinding{}),
    "Workflow" : reflect.TypeOf(Workflow{}),
    "InputRecordField" : reflect.TypeOf(InputRecordField{}),
    "SchemaDefRequirement" : reflect.TypeOf(SchemaDefRequirement{}),
    "CommandOutputEnumSchema" : reflect.TypeOf(CommandOutputEnumSchema{}),
    "ArraySchema" : reflect.TypeOf(ArraySchema{}),
    "WorkflowOutputParameter" : reflect.TypeOf(WorkflowOutputParameter{}),
    "RecordField" : reflect.TypeOf(RecordField{}),
    "InlineJavascriptRequirement" : reflect.TypeOf(InlineJavascriptRequirement{}),
    "RecordSchema" : reflect.TypeOf(RecordSchema{}),
    "CommandInputRecordSchema" : reflect.TypeOf(CommandInputRecordSchema{}),
    "OutputParameter" : reflect.TypeOf(OutputParameter{}),
    "ExpressionTool" : reflect.TypeOf(ExpressionTool{}),
    "CommandOutputBinding" : reflect.TypeOf(CommandOutputBinding{}),
    "CommandLineTool" : reflect.TypeOf(CommandLineTool{}),
    "CommandOutputParameter" : reflect.TypeOf(CommandOutputParameter{}),
    "EnvironmentDef" : reflect.TypeOf(EnvironmentDef{}),
    "OutputRecordSchema" : reflect.TypeOf(OutputRecordSchema{}),
    "InputEnumSchema" : reflect.TypeOf(InputEnumSchema{}),
    "InputArraySchema" : reflect.TypeOf(InputArraySchema{}),
    "WorkflowStep" : reflect.TypeOf(WorkflowStep{}),
    "CommandOutputArraySchema" : reflect.TypeOf(CommandOutputArraySchema{}),
    "CommandOutputRecordField" : reflect.TypeOf(CommandOutputRecordField{}),
    "File" : reflect.TypeOf(File{}),
    "InputParameter" : reflect.TypeOf(InputParameter{}),
    "OutputRecordField" : reflect.TypeOf(OutputRecordField{}),
    "CommandOutputRecordSchema" : reflect.TypeOf(CommandOutputRecordSchema{}),
    "CommandInputEnumSchema" : reflect.TypeOf(CommandInputEnumSchema{}),
    "CommandInputParameter" : reflect.TypeOf(CommandInputParameter{}),
}
type CommandInputRecordField struct {
    Doc string   `json:"doc"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Name string   `json:"name"`
    Label string   `json:"label"`
}

type OutputEnumSchema struct {
    Symbols []string   `json:"symbols"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Label string   `json:"label"`
}

type CommandInputArraySchema struct {
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Label string   `json:"label"`
}

type EnumSchema struct {
    Symbols []string   `json:"symbols"`
}

type ExpressionToolOutputParameter struct {
    Streamable bool   `json:"streamable"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Id string   `json:"id"`
    Label string   `json:"label"`
}

func (self *ExpressionToolOutputParameter) GetID() string {
    return self.Id
}

type WorkflowStepInput struct {
    Default interface{}   `json:"default"`
    LinkMerge LinkMergeMethod   `json:"linkMerge"`
    Id string   `json:"id"`
}

func (self *WorkflowStepInput) GetID() string {
    return self.Id
}

type InputRecordSchema struct {
    Fields []RecordField   `json:"fields"`
    Label string   `json:"label"`
}

type WorkflowStepOutput struct {
    Id string   `json:"id"`
}

func (self *WorkflowStepOutput) GetID() string {
    return self.Id
}

type OutputArraySchema struct {
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Label string   `json:"label"`
}

type CommandLineBinding struct {
    ShellQuote bool   `json:"shellQuote"`
    LoadContents bool   `json:"loadContents"`
    Separate bool   `json:"separate"`
    Prefix string   `json:"prefix"`
    ItemSeparator string   `json:"itemSeparator"`
    Position int   `json:"position"`
}

type Workflow struct {
    CwlVersion CWLVersion   `json:"cwlVersion"`
    Inputs []InputParameter   `json:"inputs"`
    Doc string   `json:"doc"`
    Label string   `json:"label"`
    Steps []WorkflowStep   `json:"steps"`
    Outputs []WorkflowOutputParameter   `json:"outputs"`
    Id string   `json:"id"`
    Class string   `json:"class"`
    Hints []interface{}   `json:"hints"`
}

func (self *Workflow) GetID() string {
    return self.Id
}

type InputRecordField struct {
    Doc string   `json:"doc"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Name string   `json:"name"`
    Label string   `json:"label"`
}

type SchemaDefRequirement struct {
    Class string   `json:"class"`
}

type CommandOutputEnumSchema struct {
    Symbols []string   `json:"symbols"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Label string   `json:"label"`
}

type ArraySchema struct {
}

type WorkflowOutputParameter struct {
    Streamable bool   `json:"streamable"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Label string   `json:"label"`
    LinkMerge LinkMergeMethod   `json:"linkMerge"`
    Id string   `json:"id"`
}

func (self *WorkflowOutputParameter) GetID() string {
    return self.Id
}

type RecordField struct {
    Doc string   `json:"doc"`
    Name string   `json:"name"`
}

type InlineJavascriptRequirement struct {
    Class string   `json:"class"`
    ExpressionLib []string   `json:"expressionLib"`
}

type RecordSchema struct {
    Fields []RecordField   `json:"fields"`
}

type CommandInputRecordSchema struct {
    Fields []RecordField   `json:"fields"`
    Label string   `json:"label"`
}

type OutputParameter struct {
    Streamable bool   `json:"streamable"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Id string   `json:"id"`
    Label string   `json:"label"`
}

func (self *OutputParameter) GetID() string {
    return self.Id
}

type ExpressionTool struct {
    CwlVersion CWLVersion   `json:"cwlVersion"`
    Inputs []InputParameter   `json:"inputs"`
    Outputs []ExpressionToolOutputParameter   `json:"outputs"`
    Label string   `json:"label"`
    Doc string   `json:"doc"`
    Id string   `json:"id"`
    Class string   `json:"class"`
    Hints []interface{}   `json:"hints"`
}

func (self *ExpressionTool) GetID() string {
    return self.Id
}

type CommandOutputBinding struct {
    LoadContents bool   `json:"loadContents"`
}

type CommandLineTool struct {
    CwlVersion CWLVersion   `json:"cwlVersion"`
    Inputs []CommandInputParameter   `json:"inputs"`
    PermanentFailCodes []int   `json:"permanentFailCodes"`
    Id string   `json:"id"`
    SuccessCodes []int   `json:"successCodes"`
    Doc string   `json:"doc"`
    Label string   `json:"label"`
    Outputs []CommandOutputParameter   `json:"outputs"`
    TemporaryFailCodes []int   `json:"temporaryFailCodes"`
    Class string   `json:"class"`
    Hints []interface{}   `json:"hints"`
}

func (self *CommandLineTool) GetID() string {
    return self.Id
}

type CommandOutputParameter struct {
    Streamable bool   `json:"streamable"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Id string   `json:"id"`
    Label string   `json:"label"`
}

func (self *CommandOutputParameter) GetID() string {
    return self.Id
}

type EnvironmentDef struct {
    EnvName string   `json:"envName"`
}

type OutputRecordSchema struct {
    Fields []RecordField   `json:"fields"`
    Label string   `json:"label"`
}

type InputEnumSchema struct {
    Symbols []string   `json:"symbols"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Label string   `json:"label"`
}

type InputArraySchema struct {
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Label string   `json:"label"`
}

type WorkflowStep struct {
    Doc string   `json:"doc"`
    Label string   `json:"label"`
    In []WorkflowStepInput   `json:"in"`
    ScatterMethod ScatterMethod   `json:"scatterMethod"`
    Id string   `json:"id"`
    Hints []interface{}   `json:"hints"`
}

func (self *WorkflowStep) GetID() string {
    return self.Id
}

type CommandOutputArraySchema struct {
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Label string   `json:"label"`
}

type CommandOutputRecordField struct {
    Doc string   `json:"doc"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Name string   `json:"name"`
}

type File struct {
    Format string   `json:"format"`
    Checksum string   `json:"checksum"`
    Basename string   `json:"basename"`
    Nameroot string   `json:"nameroot"`
    Nameext string   `json:"nameext"`
    Location string   `json:"location"`
    Path string   `json:"path"`
    Dirname string   `json:"dirname"`
    Contents string   `json:"contents"`
    Size int64   `json:"size"`
}

type InputParameter struct {
    Default interface{}   `json:"default"`
    Streamable bool   `json:"streamable"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Id string   `json:"id"`
    Label string   `json:"label"`
}

func (self *InputParameter) GetID() string {
    return self.Id
}

type OutputRecordField struct {
    Doc string   `json:"doc"`
    OutputBinding CommandOutputBinding   `json:"outputBinding"`
    Name string   `json:"name"`
}

type CommandOutputRecordSchema struct {
    Fields []RecordField   `json:"fields"`
    Label string   `json:"label"`
}

type CommandInputEnumSchema struct {
    Symbols []string   `json:"symbols"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Label string   `json:"label"`
}

type CommandInputParameter struct {
    Default interface{}   `json:"default"`
    Streamable bool   `json:"streamable"`
    InputBinding CommandLineBinding   `json:"inputBinding"`
    Id string   `json:"id"`
    Label string   `json:"label"`
}

func (self *CommandInputParameter) GetID() string {
    return self.Id
}

