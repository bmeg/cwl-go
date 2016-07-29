
package cwl

import (
  //"fmt"
  "log"
  "io"
  "io/ioutil"
  "encoding/json"
  "gopkg.in/yaml.v2"
)

func InputParse(path string) (JSONDict, error) {
  source, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, err
  }
  doc := make(JSONDict)
  err = yaml.Unmarshal(source, &doc)
  return doc, err
}

type JSONDict map[string]interface{}

func (self *JSONDict) Write(o io.Writer) {
  jout, err := json.Marshal(self)
  if err != nil {
    return
  }
  o.Write(jout)
}

type CWLDocData map[string]interface{} 

func Parse(path string) CWLDoc {
  source, err := ioutil.ReadFile(path)
  if err != nil {
    panic(err)
  }
  doc := make(CWLDocData)
  err = yaml.Unmarshal(source, &doc)
  
  if doc["class"].(string) == "CommandLineTool" {
    return NewCommandLineTool(doc)
  }
  return nil
}


type CWLDoc interface {
  
}

type CommandLineTool struct {
  Id string
  Inputs map[string]CommandInput
  Outputs map[string]CommandOutput
  BaseCommand []string
  Requirements []Requirement
}

func NewCommandLineTool(doc CWLDocData) CWLDoc {
  log.Printf("CommandLineTool")
  out := CommandLineTool{}
  
  if _, ok := doc["id"]; ok {
    out.Id = doc["id"].(string)
  } else {
    out.Id = ""
  }
  
  if base, ok := doc["baseCommand"].([]string); ok {
    out.BaseCommand = base
  } else {
    if base, ok := doc["baseCommand"].(string); ok {
      out.BaseCommand = []string{base}
    }
  }
  return out
}

func (self *CommandLineTool) Evaluate(inputs JSONDict) []string {
  out := make([]string, 0)
  out = append(out, self.BaseCommand...)
  return out
}


type CommandInput struct {}

type CommandOutput struct {}

type Requirement struct {}

