package cwl_engine

import (
	"fmt"
	"cwl"
	//"encoding/json"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"crypto/sha1"
	"io"
)

func NewLocalRunner(config Config) CWLRunner {
	return LocalRunner{Config: config}
}

type LocalRunner struct {
	Config Config
}

func (self LocalRunner) LocationToPath(location string) string {
	return location
}

func (self LocalRunner) RunCommand(job cwl.Job) (cwl.JSONDict, error) {
	log.Printf("Command Files: %#v", job.Files)
	log.Printf("Command Inputs: %#v", job.Inputs)

	inputs := MapInputs(job.Inputs, self)

	workdir, err := ioutil.TempDir(self.Config.TmpdirPrefix, "cwlwork_")
	if err != nil {
		return cwl.JSONDict{}, fmt.Errorf("Unable to create working dir")
	}
	log.Printf("Command Args: %s", job.Cmd)

	cmd_args := make([]string, len(job.Cmd))
	for i := range job.Cmd {
		cmd_args[i], _ = cwl.ExpressionEvaluate(job.Cmd[i], inputs)
	}

	cmd := exec.Command(cmd_args[0], cmd_args[1:]...)

	if job.Stdout != "" {
		stdout, _ := cwl.ExpressionEvaluate(job.Stdout, inputs)
		cmd.Stdout, _ = os.Create(filepath.Join(workdir, stdout))
	}
	if job.Stderr != "" {
		stderr, _ := cwl.ExpressionEvaluate(job.Stderr, inputs)
		cmd.Stderr, _ = os.Create(filepath.Join(workdir, stderr))
	}
	if job.Stdin != "" {
		stdin, _ := cwl.ExpressionEvaluate(job.Stdin, inputs)
		log.Printf("STDIN: %s", stdin)
		var err error
		cmd.Stdin, err = os.Open(stdin)
		if err != nil {
			return cwl.JSONDict{}, err
		}
	}
	cmd.Dir = workdir
	log.Printf("Workdir: %s", workdir)
	cmd_err := cmd.Run()

	meta := map[interface{}]interface{}{}
	if _, err := os.Stat(filepath.Join(workdir, "cwl.output.json")); !os.IsNotExist(err) {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(filepath.Join(workdir, "cwl.output.json"))
		//err := json.Unmarshal(data, &out)
		err := yaml.Unmarshal(data, &meta)
		log.Printf("Returned: %s = %s %s", data, meta, err)
		return meta, nil
	}

	out := cwl.JSONDict{}
	for _, o := range job.Files {
		if o.Output {
			if o.Glob != "" { //BUG: should actually type check the schema declaration here
				log.Printf("Output %s", filepath.Join(workdir, o.Glob))
				g, _ := filepath.Glob(filepath.Join(workdir, o.Glob))
				for _, p := range g {
					log.Printf("Found %s", p)
					hasher := sha1.New()
					file, _ := os.Open(p)
					if _, err := io.Copy(hasher, file); err != nil {
    				log.Fatal(err)
					}
					hash_val := fmt.Sprintf("sha1$%x", hasher.Sum([]byte{}))
					file.Close()
					info, _ := os.Stat(p)
					f := cwl.JSONDict{"location": p, "checksum" : hash_val, "class": "File", "size" : info.Size()}
					out[o.Id] = f
				}
			} else {
				if a, ok := meta[o.Id]; ok {
					out[o.Id] = a
				}
			}
		}
	}
	return out, cmd_err
}
