package main

import (
  "os"
  "fmt"
  "flag"
  "cwl"
  "log"
  "cwl/engine"
  "encoding/json"
)



func main() {
  var version_flag = flag.Bool("version", false, "version")
  var tmp_outdir_prefix = flag.String("tmp-outdir-prefix", "", "Temp output prefix")
  var tmpdir_prefix = flag.String("tmpdir-prefix", "", "Tempdir prefix")
  var outdir = flag.String("outdir", "./", "Outdir")
  var quiet_flag = flag.Bool("quiet", false, "quiet")
  flag.Parse()

  if (*version_flag) {
    fmt.Printf("cwlgo-tool v0.0.1\n")
    return
  }
  fmt.Fprintf(os.Stderr, "cwlgo-tool v0.0.1\n")

  config := cwl.Config{ 
    TmpOutdirPrefix: *tmp_outdir_prefix,
    TmpdirPrefix: *tmpdir_prefix,
    Outdir: *outdir,
    Quiet: *quiet_flag,
  }

  cwl_doc := cwl.Parse( flag.Arg(0) )
  inputs, _ := cwl.InputParse( flag.Arg(1) )

  runner := cwl_engine.NewLocalRunner(config)
  
  graphState := cwl_doc.NewGraphState(inputs)
  for !cwl_doc.Done(graphState) {
    for _, step := range cwl_doc.ReadySteps(graphState) {
      job, err := cwl_doc.GenerateJob(step, graphState)
      if err != nil {
        log.Printf("%s", err)
        return
      }
      out, err := runner.RunCommand(job)
      if err != nil {
        log.Printf("%s", err)
        return
      }
      graphState = cwl_doc.UpdateStepResults(graphState, step, out)
    }
  }
  out := cwl_doc.GetResults(graphState)
  o, _ := json.Marshal(out)
  fmt.Printf(string(o))
  
}