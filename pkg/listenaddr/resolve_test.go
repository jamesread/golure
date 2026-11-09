package listenaddr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name       string
		port       string
		configured string
		fallback   string
		want       string
	}{
		{"configured without PORT", "", ":9090", ":8080", ":9090"},
		{"fallback without PORT", "", "", ":8080", ":8080"},
		{"blank configured uses fallback", "", "  ", ":8080", ":8080"},
		{"bare port keeps empty host", "3000", ":9090", ":8080", ":3000"},
		{"bare port keeps configured host", "9000", "127.0.0.1:1337", "0.0.0.0:1", "127.0.0.1:9000"},
		{"bare port uses fallback host", "9000", "", "0.0.0.0:8080", "0.0.0.0:9000"},
		{"full address", "0.0.0.0:4000", ":9090", ":8080", "0.0.0.0:4000"},
		{"colon port", ":4000", "127.0.0.1:9090", ":8080", ":4000"},
		{"whitespace trimmed", " 3000 ", ":9090", ":8080", ":3000"},
		{"non numeric", "not-a-port", "0.0.0.0:1337", ":8080", "0.0.0.0:1337"},
		{"above max", "65536", "0.0.0.0:1337", ":8080", "0.0.0.0:1337"},
		{"negative", "-1", "0.0.0.0:1337", ":8080", "0.0.0.0:1337"},
		{"zero", "0", "0.0.0.0:1337", ":8080", "0.0.0.0:1337"},
		{"invalid full address", "host:notaport", ":9090", ":8080", ":9090"},
		{"malformed configured keeps configured", "9000", "nonsense", ":8080", "nonsense"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PORT", tt.port)

			assert.Equal(t, tt.want, Resolve(tt.configured, tt.fallback))
		})
	}
}
