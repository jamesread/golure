package easyexec

import (
	"bytes"
	"time"
	"context"
	"os/exec"
	"os"
	log "github.com/sirupsen/logrus"
)

type ExecResult struct {
	Output string
	ExitCode int
	Error error
	WorkingDirectory string
}

type ExecRequest struct {
	Executable string
	Args []string
	WorkingDirectory string
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10) * time.Second)
	defer cancel()

	streamer := &OutputStreamer{Output: &bytes.Buffer{}}

	if wd == "" {
		var err error
		wd, err = os.Getwd()

		if err != nil {
			return &ExecResult{Error: err}
		}	
	}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = wd
	cmd.Stdout = streamer;
	cmd.Stderr = streamer;

	runerr := cmd.Run()

	return &ExecResult{
		Output: streamer.String(),
		Error: runerr,
		ExitCode: cmd.ProcessState.ExitCode(),
		WorkingDirectory: wd,
	}
}

func ExecWithRequest(req *ExecRequest) (*ExecResult) {	
	return Exec(req.Executable, req.Args, req.WorkingDirectory)
}

func ExecWithReqLog(req *ExecRequest) (*ExecResult) {
	log.Infof("cmd: %v %v", req.Executable, req.Args)
	log.Infof("wd: %v", req.WorkingDirectory)

	ret := Exec(req.Executable, req.Args, req.WorkingDirectory)

	if ret.Error != nil {
		log.Errorf("err: %v", ret.Error)
	}

	log.Infof("stdout: %v", ret.Output)

	return ret
}
