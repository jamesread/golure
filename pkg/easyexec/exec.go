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

func ExecShell(executable string) (*ExecResult) {
	return Exec("sh", []string { "-c", executable })
}

func Exec(executable string, args []string) (*ExecResult) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10) * time.Second)
	defer cancel()

	streamer := &OutputStreamer{Output: &bytes.Buffer{}}

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdout = streamer;
	cmd.Stderr = streamer;

	runerr := cmd.Run()

	return &ExecResult{
		Output: streamer.String(),
		Error: runerr,
		ExitCode: cmd.ProcessState.ExitCode(),
	}
}

func ExecLog(executable string, args []string) (*ExecResult) {
	cwd, _ := os.Getwd()

	log.Infof("cwd: %v", cwd)
	log.Infof("cmd: %v %v", executable, args)

	ret := Exec(executable, args)

	if ret.Error != nil {
		log.Errorf("err: %v", ret.Error)
	}

	log.Infof("stdout: %v", ret.Output)

	return ret
}
