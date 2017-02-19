
package cwlparser


import (
  "log"
  "fmt"
  "io/ioutil"
  "reflect"
  //"path/filepath"
  "gopkg.in/yaml.v2"
)

type JSONDict map[interface{}]interface{}

type CWLGraph struct {
	Elements map[string]CWLClass
	Main     string
}

func (graph *CWLGraph) GetID() string {
  return graph.Main
}

type Schema interface {

}

type CWLParser struct {
	Path     string
	Schemas  map[string]Schema
	Elements map[string]CWLClass
}


type CWLClass interface {
	GetID() string
}


func NewParser(path string) *CWLParser {
  return &CWLParser{Path: path, Schemas: make(map[string]Schema), Elements: make(map[string]CWLClass)}
}


func (parser *CWLParser) Parse() (CWLClass, error) {
	source, err := ioutil.ReadFile(parser.Path)
	if err != nil {
		return &CWLGraph{}, fmt.Errorf("Unable to parse file: %s", err)
	}
  doc := make(map[interface{}]interface{})
  err = yaml.Unmarshal(source, &doc)
  //x, _ := filepath.Abs(cwl_path)

  if base, ok := doc["$graph"]; ok {
    return parser.NewGraph(base)
  } else if _, ok := doc["class"]; ok {
    return NewDocClass(doc)
  }
  return &CWLGraph{}, fmt.Errorf("Unable to parse file")
}

func NewClass(doc JSONDict, t reflect.Type) (CWLClass, error) {
  log.Printf("New Class: %s", t)
  o := reflect.New(t)
  for k, v := range doc {
    log.Printf("%s = %s", k, v)
  }
  ov := o.Interface()
  return ov.(CWLClass), nil
}

func NewDocClass(doc JSONDict) (CWLClass, error) {
  //fmt.Printf("%s\n", doc["class"])
  /*
	if doc["class"].(string) == "Workflow" {
		return self.NewWorkflow(doc), nil
	} else if doc["class"].(string) == "CommandLineTool" {
		return self.NewCommandLineTool(doc), nil
	} else if doc["class"].(string) == "ExpressionTool" {
		return self.NewExpressionTool(doc), nil
	}
  */

  t := TYPES[doc["class"].(string)]
  o := reflect.New(t)
  for i := 0; i < t.NumField(); i++ {
    f := t.Field(i)

    if v, ok := doc[ f.Tag.Get("json") ]; ok {
      //fmt.Printf("%s %#v\n", f.Name, v )
      if f.Type.Kind() == reflect.String {
        o.Elem().Field(i).SetString(v.(string))
      } else if f.Type.Kind() == reflect.Slice {
        log.Printf("Loading: %s into %s", v, f.Type.Elem())
        if x, ok := v.(map[interface{}]interface{}); ok {
          log.Printf("translate: %s", x)
          for el_key, el_val := range x {
            if el_val_map, ok := el_val.(map[interface{}]interface{}); ok {
              if f.Type.Elem().Kind() == reflect.Struct {
                el, _ := NewClass(el_val_map, f.Type.Elem())
                log.Printf("element %s %s", el_key, el)
                ea := reflect.Append( o.Elem().Field(i), reflect.ValueOf(el).Elem() )
                o.Elem().Field(i).Set(ea)
              }
            }
          }
        } else if x , ok := v.([]interface{}); ok {
          for _, el_val := range x {
            if el_val_map, ok := el_val.(map[interface{}]interface{}); ok {
              if f.Type.Elem().Kind() == reflect.Struct {
                el, _ := NewClass(el_val_map, f.Type.Elem())
                log.Printf("element %s", el)
                ea := reflect.Append( o.Elem().Field(i), reflect.ValueOf(el).Elem() )
                o.Elem().Field(i).Set(ea)
              }
            }
          }
        } else {
          log.Printf("unknown: %#v", v)
        }
      } else {
        log.Printf("Unknown type: %s", f.Type.Kind())
      }
    }
  }
  ov := o.Interface()
	return ov.(CWLClass), nil
}

func (self *CWLParser) NewGraph(graph interface{}) (*CWLGraph, error) {
	docs := CWLGraph{Elements: map[string]CWLClass{}}
	if base, ok := graph.([]interface{}); ok {
		for _, i := range base {
			//parser := CWLParser{Path: self.Path, Schemas: make(map[string]Schema), Elements: make(map[string]CWLClass)}
			if classBase, ok := i.(map[interface{}]interface{}); ok {
				cDoc, err := NewDocClass(classBase)
				if err != nil {
					return &docs, err
				}
				/*
        for k, v := range cDoc.Elements {
					docs.Elements[k] = v
				}
        */
				log.Printf("Parsing Graph %#v", cDoc)
			}
		}
	}
	return &docs, nil
}
