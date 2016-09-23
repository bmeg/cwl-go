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
	"strings"
	"time"
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

	cwl_path := flag.Arg(0)
	element_id := ""
	if strings.Contains(cwl_path, "#") {
		tmp := strings.Split(cwl_path, "#")
		cwl_path = tmp[0]
		element_id = tmp[1]
	}
	cwl_docs, err := cwl.Parse(cwl_path)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Unable to parse CWL document: %s\n", err))
		if _, ok := err.(cwl.UnsupportedRequirement); ok {
			os.Exit(33)
		}
		os.Exit(1)
	}
	log.Printf("CWLDoc: %#v", cwl_docs)
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

	if cwl_docs.Main == "" {
		if element_id == "" {
			os.Stderr.WriteString(fmt.Sprintf("Need to define element ID\n"))
			os.Exit(1)
		}
		cwl_docs.Main = element_id
	}

	cwl_doc := cwl_docs.Elements[cwl_docs.Main]
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
			var runner cwl_engine.JobRunner
			if job.JobType == cwl.EXPRESSION {
				runner = cwl_engine.NewExpressionRunner(config)
			} else {
				if job.DockerImage != "" {
					runner, _ = cwl_engine.NewDockerRunner(config)
				} else {
					runner, _ = cwl_engine.NewLocalRunner(config)
				}
			}

			task, err := cwl_engine.StartJob(job, runner)
			if err != nil {
				log.Printf("Runtime Error: %s", err)
				os.Exit(1)
			}

			sleepTime := time.Microsecond
			for !cwl_engine.JobDone(task, runner) {
				time.Sleep(sleepTime)
				if sleepTime < time.Second*10 {
					sleepTime += time.Millisecond
				}
			}

			out, _ := cwl_engine.CleanupJob(task, runner)
			graphState = cwl_doc.UpdateStepResults(graphState, step, out)
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
