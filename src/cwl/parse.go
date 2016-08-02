
package cwl

import (
  "fmt"
  "log"
  "io"
  "sort"
  "strings"
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
  Arguments []Argument
}

func NewCommandLineTool(doc CWLDocData) CWLDoc {
  log.Printf("CommandLineTool: %v", doc)
  out := CommandLineTool{}
  out.Inputs = make(map[string]CommandInput)
  out.Outputs = make(map[string]CommandOutput)
  
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
  
  if base, ok := doc["arguments"]; ok {
    for _, x := range base.([]interface{}) {
      n, err := NewArgument(x)
      if err == nil {
        out.Arguments = append(out.Arguments, n )
      }
    }
  }
  
  if base, ok := doc["inputs"]; ok {    
    if base_map, ok := base.(map[string]interface{}); ok {
      for k, v := range base_map {
        n, err := NewCommandInput(v)
        if err == nil {
          n.Id = k
          out.Inputs[n.Id] = n
        }
      }
    } else if base_array, ok := base.([]interface{}); ok {
      log.Printf("Input array: %d", len(base_array))
      for _, x := range base_array {
        n, err := NewCommandInput(x)
        if err == nil {
          out.Inputs[n.Id] = n
        } else {
          log.Printf("Command line Input error: %s", err)
        }
      }
    } else {
      log.Printf("Can't Parse Inputs")
    }
  } else {
    log.Printf("No Inputs found")
  }
  
  log.Printf("Parse CommandLineTool: %v", out)
  return out
}

type cmdArg struct {
  position int
  value []string
}

type cmdArgArray []cmdArg


func (self cmdArgArray) Len() int {
  return len(self)
}

func (self cmdArgArray) Less(i, j int) bool {
  return (self)[i].position < (self)[j].position
}

func (self cmdArgArray) Swap(i, j int) {
  (self)[i], (self)[j] = (self)[j], (self)[i]
}

func (self *CommandLineTool) Evaluate(inputs JSONDict) []string {
  log.Printf("CommandLineTool Evalute")
  out := make([]string, 0)
  out = append(out, self.BaseCommand...)
  
  args := make(cmdArgArray, 0, len(self.Arguments) + len(self.Inputs))
  //Arguments
  for _, x := range self.Arguments {
    c := cmdArg{
      position: x.Position,
      value: x.Evaluate(inputs),
    }
    args = append(args, c)
  }  
  //Inputs
  for _, x := range self.Inputs {
    c := cmdArg{
      position: x.Position,
      value: x.Evaluate(inputs),
    }
    args = append(args, c)
  }  

  sort.Stable(args)
  for _, x := range args {
    out = append(out, x.value...)
  }
  log.Printf("Out: %v", out)
  return out
}


type CommandInput struct {
  Id string
  Position int
  Prefix *string
  ItemSeparator *string
  Type DataType
}

func NewCommandInput(x interface{}) (CommandInput, error) {
  out := CommandInput{}
  if base, ok := x.(map[interface{}]interface{}); ok {
    out.Id = base["id"].(string)
    if binding, ok := base["inputBinding"]; ok {
      if pos, ok := binding.(map[interface{}]interface{})["position"]; ok {
        out.Position = pos.(int)
      } else {
        out.Position = 100000
      }
      if prefix, ok := binding.(map[interface{}]interface{})["prefix"].(string); ok {
        out.Prefix = &prefix
      }
      if itemSep, ok := binding.(map[interface{}]interface{})["itemSeparator"].(string); ok {
        out.ItemSeparator = &itemSep
      }
    }
    out.Type = NewDataType( base["type"] )
  } else {
    return out, fmt.Errorf("Unable to parse CommandInput: %v", x)
  }
  log.Printf("CommandInput: %#v", out)
  return out, nil
}


type DataType struct {
  TypeName string
  Items *DataType
}

func NewDataType(value interface{}) DataType {
    if base, ok := value.(string); ok {
      return DataType{TypeName:base}
    } else if base, ok := value.(map[interface{}]interface{}); ok {
      out := DataType{TypeName:base["type"].(string)}
      a := NewDataType(base["items"])
      out.Items = &a
      return out
    } else {      
      panic(fmt.Sprintf("Unknown data type: %#v\n", value))
    }
    return DataType{}
}

func (self *DataType) Evaluate(value interface{}) []string {
  if self.TypeName == "File" {
    if base, ok := value.(map[interface{}]interface{}); ok {
      if class, ok := base["class"]; ok {
        if class.(string) == "File" {
          return []string{ base["location"].(string) }
        }
      }
    }
  } else if self.TypeName == "int" {
      return []string{fmt.Sprintf("%d", value.(int))}
  } else if self.TypeName == "array" {
      out := []string{}
      if base, ok := value.([]interface{}); ok {
        for _, i := range base {
          out = append(out, self.Items.Evaluate(i)...)
        }      
      }
      return out
  } else {
    panic(fmt.Sprintf("Unknown Type %s", self.TypeName))
  }

  return []string{}
}

func (self *CommandInput) Evaluate(inputs JSONDict) []string {
  out := []string{}

  if base, ok := inputs[self.Id]; ok {
    value_str := self.Type.Evaluate(base)
    if self.ItemSeparator != nil {
      value_str = []string{ strings.Join( value_str, *self.ItemSeparator ) }
    }
    if self.Prefix != nil {
      value_str = append( []string{*self.Prefix}, value_str... )
    }
    return value_str
  } else {
    return []string{}
  }
  
  return out
}


type CommandOutput struct {}

type Requirement struct {}


type Argument struct {
  Value *string
  ValueFrom *string
  Position int
  Prefix *string
  ItemSeparator *string
}

func NewArgument(x interface{}) (Argument, error) {
  if base, ok := x.(string); ok {
    return Argument{Value:&base}, nil
  }
  if base, ok := x.(map[interface{}]interface{}); ok {
    out := Argument{}
    if x, ok := base["valueFrom"]; ok {
        s := x.(string)
        out.ValueFrom = &s
    }
    if x, ok := base["position"]; ok {
        out.Position = x.(int)
    } else {
        out.Position = 10000
    }
    if x, ok := base["prefix"]; ok {
        x_s := x.(string)
        out.Prefix = &x_s
    }
    return out, nil
  }
  return Argument{}, fmt.Errorf("Can't Parse Argument")
}

func (self *Argument) Evaluate(inputs JSONDict) []string {
  //fmt.Printf("Arguments %#v\n", self)
  if self.Value != nil {
    return []string{*self.Value}
  } else if self.ValueFrom != nil {
    //fmt.Println("Arguments Evaluate")
    exp, _ := ExpressionEvaluate(*self.ValueFrom, inputs)
    value_str := []string{ exp }
    if self.ItemSeparator != nil {
      value_str = []string{ strings.Join( value_str, *self.ItemSeparator ) }
    }
    if self.Prefix != nil {
      value_str = append( []string{*self.Prefix}, value_str... )
    }
    return value_str
  }
  return []string{}
}
