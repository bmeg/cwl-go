package main

import (
	//"os"
	"cwl"
	"cwl/engine"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	var version_flag = flag.Bool("version", false, "version")
	var tmp_outdir_prefix_flag = flag.String("tmp-outdir-prefix", "./", "Temp output prefix")
	var tmpdir_prefix_flag = flag.String("tmpdir-prefix", "/tmp", "Tempdir prefix")
	var outdir = flag.String("outdir", "./", "Outdir")
	var quiet_flag = flag.Bool("quiet", false, "quiet")
	flag.Parse()

	if *version_flag {
		fmt.Printf("cwlgo-tool v0.0.1\n")
		return
	}
	//fmt.Fprintf(os.Stderr, "cwlgo-tool v0.0.1\n")

	if *quiet_flag {
		log.SetOutput(ioutil.Discard)
	}

	tmp_outdir_prefix, _ := filepath.Abs(*tmp_outdir_prefix_flag)
	tmpdir_prefix, _ := filepath.Abs(*tmpdir_prefix_flag)

	config := cwl_engine.Config{
		TmpOutdirPrefix: tmp_outdir_prefix,
		TmpdirPrefix:    tmpdir_prefix,
		Outdir:          *outdir,
		Quiet:           *quiet_flag,
	}

	cwl_doc := cwl.Parse(flag.Arg(0))
	inputs, _ := cwl.InputParse(flag.Arg(1))

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
