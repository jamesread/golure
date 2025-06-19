package redact

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactStrings(t *testing.T) {
	testStrings := []struct {
		Name string
		Input string
		Expected string
	}{
		{
			Name: "Simple Redaction",
			Input: "hello world",
			Expected: "****orld",
		},
		{
			Name: "Short String Redaction",
			Input: "hi",
			Expected: "****",
		},
		{
			Name: "Exact Length Redaction",
			Input: "test",
			Expected: "****",
		},
	}

	for _, test := range testStrings {
		t.Run(test.Name, func(t *testing.T) {
			redacted := RedactString(test.Input)
			assert.Equal(t, test.Expected, redacted, "Expected redacted string to match")
		});
	}
}
