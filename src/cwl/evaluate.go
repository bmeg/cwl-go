
package cwl


func Evaluate(doc CWLDoc, inputs JSONDict) JSONDict {
  out := make(JSONDict)
  if cmd, ok := doc.(CommandLineTool); ok {
    cmd_line := cmd.Evaluate(inputs)
    out["args"] = cmd_line[1:]
  }
  return out
}