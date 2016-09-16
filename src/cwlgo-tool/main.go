package main

import (
	"cwl"
	"cwl/engine"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	cwl_doc, err := cwl.Parse(flag.Arg(0))
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Unable to parse CWL document: %s\n", err))
		if _, ok := err.(cwl.UnsupportedRequirement); ok {
			os.Exit(33)
		}
		os.Exit(1)
	}
	log.Printf("CWLDoc: %#v", cwl_doc)
	var inputs cwl.JSONDict
	if len(flag.Args()) == 1 {
		inputs = cwl.JSONDict{}
	} else {
		var err error
		inputs, err = cwl.InputParse(flag.Arg(1))
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("Unable to parse Input document: %s\n", err))
			os.Exit(1)
		}
	}
	runner := cwl_engine.NewLocalRunner(config)
	expression_runner := cwl_engine.NewExpressionRunner(config)
	log.Printf("Starting run")
	graphState := cwl_doc.NewGraphState(inputs)
	for !cwl_doc.Done(graphState) {
		readyCount := 0
		for _, step := range cwl_doc.ReadySteps(graphState) {
			job, err := cwl_doc.GenerateJob(step, graphState)
			if err != nil {
				log.Printf("%s", err)
				os.Exit(1)
			}
			if job.JobType == cwl.EXPRESSION {
				out, err := expression_runner.RunCommand(job)
				if err != nil {
					log.Printf("Runtime Error: %s", err)
					os.Exit(1)
				}
				graphState = cwl_doc.UpdateStepResults(graphState, step, out)
			} else {
				out, err := runner.RunCommand(job)
				if err != nil {
					log.Printf("Runtime Error: %s", err)
					os.Exit(1)
				}
				graphState = cwl_doc.UpdateStepResults(graphState, step, out)
			}
			readyCount += 1
		}
		if readyCount == 0 {
			log.Printf("No jobs found")
			return
		}
	}
	out := cwl_doc.GetResults(graphState)
	log.Printf("doc results: %#v", out)
	fmt.Printf("%s\n", string(out.ToString()))

}
