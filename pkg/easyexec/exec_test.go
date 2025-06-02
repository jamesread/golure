package easyexec

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	ret := ExecShell("date")

	assert.NotNil(t, ret, "Expected ExecShell to return a result")
	assert.Equal(t, 0, ret.ExitCode, "Expected exit code to be 0")
	assert.NotEmpty(t, ret.Output, "Expected output to not be empty")
	assert.NoError(t, ret.Error, "Expected no error from ExecShell")
}
