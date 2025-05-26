package dirs

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetFirstExistingDirectory(t *testing.T) {
	existingDir := "/etc"

	directories := []string{"/foo/bar", existingDir}

	foundDir, err := GetFirstExistingDirectory(directories)

	assert.Equal(t, existingDir, foundDir, "Expected to find the existing directory")
	assert.Nil(t, err, "Expected no error when finding an existing directory")
}
