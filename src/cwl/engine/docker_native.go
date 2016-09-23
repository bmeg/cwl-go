package cwl_engine

import (
	"cwl"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func NewNativeDockerRunner(config Config) (JobRunner, error) {
	workdir, err := ioutil.TempDir(config.TmpdirPrefix, "cwlwork_")
	if err != nil {
		return nil, fmt.Errorf("Unable to create working dir")
	}
	return DockerNativeRunner{Config: config, hostWorkDir: workdir, fileMap: map[string]string{}, fileMapRev: map[string]string{}}, nil
}

type DockerNativeRunner struct {
	Config      Config
	hostWorkDir string
	fileMap     map[string]string
	fileMapRev  map[string]string
}

func (self DockerNativeRunner) getClient() (*client.Client, error) {
	client, err := client.NewEnvClient()
	if err != nil {
		log.Printf("Docker Error\n")
		return nil, err
	}
	return client, nil
}

func (self DockerNativeRunner) prepImage(imageName string) error {
	client, err := self.getClient()
	if err != nil {
		return err
	}
	list, err := client.ImageList(context.Background(), types.ImageListOptions{MatchName: imageName})

	if err != nil || len(list) == 0 {
		log.Printf("Image %s not found: %s", imageName, err)
		pull_opt := types.ImagePullOptions{}
		r, err := client.ImagePull(context.Background(), imageName, pull_opt)
		if err != nil {
			log.Printf("Image not pulled: %s", err)
			return err
		}
		for {
			l := make([]byte, 1000)
			_, e := r.Read(l)
			if e == io.EOF {
				break
			}
			log.Printf("%s", l)
		}
		r.Close()
		log.Printf("Image %s Pulled", imageName)
	}
	return nil
}

func (self DockerNativeRunner) GetWorkDirPath() string {
	return "/var/run/cwl-go"
}

func (self DockerNativeRunner) LocationToPath(location string) string {
	if out, ok := self.fileMapRev[location]; ok {
		return out
	}
	base := filepath.Base(location)
	out := fmt.Sprintf("/var/run/cwlinput/%d/%s", len(self.fileMap), base)
	self.fileMap[out] = location
	self.fileMapRev[location] = out
	log.Printf("Translating: %s to %s", location, out)
	return out
}

func (self DockerNativeRunner) StartProcess(inputs cwl.JSONDict, cmd_args []string, workdir, stdout, stderr, stdin, dockerImage string) (cwl.JSONDict, error) {

	self.prepImage(dockerImage)
	client, err := self.getClient()
	if err != nil {
		return cwl.JSONDict{}, err
	}
	binds := []string{fmt.Sprintf("%s:%s", self.hostWorkDir, workdir)}

	log.Printf("Docker files: %s", inputs.GetFilePaths())

	for _, n := range inputs.GetFilePaths() {
		binds = append(binds, fmt.Sprintf("%s:%s", self.fileMap[n], n))
	}
	log.Printf("Docker Binds: %s", binds)

	container, err := client.ContainerCreate(context.Background(),
		&container.Config{Cmd: cmd_args, Image: dockerImage, WorkingDir: workdir, Tty: true},
		&container.HostConfig{Binds: binds},
		&network.NetworkingConfig{},
		"",
	)

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return cwl.JSONDict{}, err
	}

	log.Printf("Starting Docker %s (mount: %s): %s", container.ID, strings.Join(binds, ","), strings.Join(cmd_args, " "))
	err = client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})

	if err != nil {
		log.Printf("Docker run Error: %s", err)
		return cwl.JSONDict{}, err
	}

	resFile := self.hostWorkDir + ".result"
	go func() {
		client, err := self.getClient()
		if err != nil {
			log.Printf("Docker client error: %s", err)
		}
		log.Printf("Attaching Container: %s", container.ID)
		exit_code, err := client.ContainerWait(context.Background(), container.ID)
		if err != nil {
			log.Printf("docker %s error: %s", container.ID, err)
		} else {
			log.Printf("docker %s complete: %s", container.ID, exit_code)
		}

		if stdout != "" {
			stdout_file, _ := os.Create(filepath.Join(self.hostWorkDir, stdout))
			stdout_log, _ := client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{ShowStdout: true})
			buffer := make([]byte, 10240)
			for {
				l, e := stdout_log.Read(buffer)
				if e == io.EOF {
					break
				}
				stdout_file.Write(buffer[:l])
			}
			stdout_file.Close()
			stdout_log.Close()
		}

		if stderr != "" {
			stderr_file, _ := os.Create(filepath.Join(self.hostWorkDir, stderr))
			stderr_log, err := client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{ShowStdout: true}) //types.ContainerLogsOptions{ShowStderr: true, Details: false})
			if err != nil {
				log.Printf("Read Error: %s", err)
			}
			buffer := make([]byte, 10240)
			for {
				l, e := stderr_log.Read(buffer)
				log.Printf("Logging: %s", buffer)
				if e == io.EOF {
					break
				}
				stderr_file.Write(buffer[:l])
			}
			stderr_log.Close()
			stderr_file.Close()
		}
		//client.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
		ioutil.WriteFile(resFile, []byte(fmt.Sprintf("%d", exit_code)), 0600)
	}()
	return cwl.JSONDict{"resFile": resFile}, nil
}

func (self DockerNativeRunner) GetOutput(prodData cwl.JSONDict) cwl.JSONDict {
	out := cwl.JSONDict{}
	path := filepath.Join(self.hostWorkDir, "cwl.output.json")
	if _, err := os.Stat(path); err == nil {
		log.Printf("Found cwl.output.json")
		data, _ := ioutil.ReadFile(path)
		err := yaml.Unmarshal(data, &out)
		log.Printf("Returned: %s = %s %s", data, out, err)
	}
	return out
}

func (self DockerNativeRunner) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(self.hostWorkDir, path))
}

func (self DockerNativeRunner) Glob(pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(self.hostWorkDir, pattern))
	return matches
}

func (self DockerNativeRunner) ExitCode(procData cwl.JSONDict) (int, bool) {
	if _, err := os.Stat(procData["resFile"].(string)); err == nil {
		d, _ := ioutil.ReadFile(procData["resFile"].(string))
		i, _ := strconv.Atoi(string(d))
		return i, true
	}
	return 0, false
}
