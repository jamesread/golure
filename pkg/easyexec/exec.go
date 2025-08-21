package easyexec

import (
	"bytes"
	"time"
	"context"
	"os/exec"
	"os"
	"math"
	log "github.com/sirupsen/logrus"
)

type ExecResult struct {
	Output string
	ExitCode int
	Error error
	WorkingDirectory string
	Timeout float64
}

type ExecRequest struct {
	Executable string
	Args []string
	WorkingDirectory string
	Timeout float64
	Log bool
}	

type OutputStreamer struct {
	Output *bytes.Buffer
}

func (s *OutputStreamer) Write(p []byte) (n int, err error) {
	return s.Output.Write(p)
}

func (s *OutputStreamer) String() string {
	return s.Output.String()
}

func ExecShell(req *ExecRequest) (*ExecResult) {
	return Exec("sh", []string { "-c", req.Executable}, req.WorkingDirectory)
}

func Exec(executable string, args []string, wd string) (*ExecResult) {
	req := &ExecRequest{
		Executable: executable,
		Args: args,
		WorkingDirectory: wd,
	}

	return ExecWithRequest(req)
}

func ExecWithRequest(req *ExecRequest) (*ExecResult) {	
	if req.Log {
		log.Infof("cmd: %v %v", req.Executable, req.Args)
		log.Infof("wd: %v", req.WorkingDirectory)
		log.Infof("timeout: %v", req.Timeout)
	}

	ret := execImpl(req)

	if (req.Log) {
		if ret.Error != nil {
			log.Errorf("err: %v", ret.Error)
		}

		log.Infof("stdout: %v", ret.Output)
		log.Infof("timeout: %v", ret.Timeout)
	}

	return ret
}

func execImpl(req *ExecRequest) (*ExecResult) {
	timeout := math.Max(10, req.Timeout)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()

	streamer := &OutputStreamer{Output: &bytes.Buffer{}}

	if req.WorkingDirectory == "" {
		var err error
		req.WorkingDirectory, err = os.Getwd()

		if err != nil {
			return &ExecResult{Error: err}
		}	
	}

	cmd := exec.CommandContext(ctx, req.Executable, req.Args...)
	cmd.Dir = req.WorkingDirectory
	cmd.Stdout = streamer;
	cmd.Stderr = streamer;

	runerr := cmd.Run()

	return &ExecResult{
		Output: streamer.String(),
		Error: runerr,
		ExitCode: cmd.ProcessState.ExitCode(),
		WorkingDirectory: req.WorkingDirectory,
		Timeout: timeout,
	}
}

