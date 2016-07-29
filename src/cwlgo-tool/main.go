package main

import (
  "os"
  "fmt"
  "flag"
  "cwl"
  "cwl/engine"
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

  if true {
    out := cwl.Evaluate(cwl_doc, inputs)
    out.Write(os.Stdout)
    os.Stdout.Write([]byte("\n"))
  } else {
    runner := cwl_engine.NewLocalRunner(config)
    runner.RunCommand(cwl_doc, inputs)
  }
  
}