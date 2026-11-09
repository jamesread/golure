package logsetup

import (
	"bytes"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		raw     string
		want    log.Level
		wantErr bool
	}{
		{"", log.WarnLevel, false},
		{"   ", log.WarnLevel, false},
		{"debug", log.DebugLevel, false},
		{" TRACE ", log.TraceLevel, false},
		{"Error", log.ErrorLevel, false},
		{"loud", log.WarnLevel, true},
	}

	for _, tt := range tests {
		got, err := ParseLevel(tt.raw, log.WarnLevel)

		assert.Equal(t, tt.want, got, tt.raw)
		assert.Equal(t, tt.wantErr, err != nil, tt.raw)
	}
}

func TestApplyEnvLevel(t *testing.T) {
	original := log.GetLevel()
	t.Cleanup(func() { log.SetLevel(original) })

	logger := log.New()

	t.Setenv("TEST_LOG_LEVEL", "debug")
	require.NoError(t, ApplyEnvLevel(logger, "TEST_LOG_LEVEL", log.InfoLevel))
	assert.Equal(t, log.DebugLevel, logger.GetLevel())
	assert.Equal(t, log.DebugLevel, log.GetLevel())

	t.Setenv("TEST_LOG_LEVEL", "loud")
	assert.Error(t, ApplyEnvLevel(nil, "TEST_LOG_LEVEL", log.InfoLevel))
	assert.Equal(t, log.InfoLevel, log.GetLevel())
}

func TestMirrorStandard(t *testing.T) {
	MirrorStandard(nil)

	original := log.GetLevel()
	t.Cleanup(func() { log.SetLevel(original) })

	log.SetLevel(log.TraceLevel)

	logger := log.New()
	logger.SetOutput(&bytes.Buffer{})
	MirrorStandard(logger)

	assert.Equal(t, log.TraceLevel, logger.GetLevel())
	assert.Equal(t, log.StandardLogger().Out, logger.Out)
}

func TestBracketFormatter(t *testing.T) {
	line, err := (&BracketFormatter{}).Format(&log.Entry{
		Level:   log.WarnLevel,
		Time:    processStart.Add(7 * time.Second),
		Message: "hello\n",
		Data:    log.Fields{"zeta": "plain", "alpha": "needs quoting", "count": 3, "empty": ""},
	})

	require.NoError(t, err)
	assert.Equal(t, "WARN[0007] hello"+string(bytes.Repeat([]byte(" "), 39))+` alpha="needs quoting" count=3 empty="" zeta=plain`+"\n", string(line))
}
