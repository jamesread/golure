package redact

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderValues(t *testing.T) {
	for _, name := range []string{"Authorization", "authorization", "AUTHORIZATION", "Proxy-Authorization", "X-Api-Key", "X-Forwarded-Access-Token", "Cookie"} {
		assert.Equal(t, []string{"[redacted]"}, HeaderValues(name, []string{"secret"}), name)
	}

	assert.Equal(t, []string{"[redacted]", "[redacted]"}, HeaderValues("set-cookie", []string{"a=1", "b=2"}))
	assert.Equal(t, []string{"https"}, HeaderValues("X-Forwarded-Proto", []string{"https"}))
}

func TestHeaderValuesDoesNotModifyInput(t *testing.T) {
	values := []string{"secret"}

	HeaderValues("Authorization", values)

	assert.Equal(t, []string{"secret"}, values)
}
