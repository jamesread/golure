// Package logsetup configures logrus the same way across services.
package logsetup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var processStart = time.Now()

// ParseLevel parses a logrus level name, ignoring case and surrounding space.
// An empty value returns fallback and no error. An invalid value returns
// fallback and the parse error.
func ParseLevel(raw string, fallback log.Level) (log.Level, error) {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return fallback, nil
	}

	level, err := log.ParseLevel(strings.ToLower(raw))

	if err != nil {
		return fallback, err
	}

	return level, nil
}

// ApplyEnvLevel sets the level named by the environment variable envVar on l,
// if l is not nil, and on the standard logger. An invalid value applies
// fallback and returns the parse error for the caller to report.
func ApplyEnvLevel(l *log.Logger, envVar string, fallback log.Level) error {
	level, err := ParseLevel(os.Getenv(envVar), fallback)

	if l != nil {
		l.SetLevel(level)
	}

	log.SetLevel(level)

	return err
}

// MirrorStandard copies the formatter, output and level of the standard logger
// to l, for loggers that must remain separate instances.
func MirrorStandard(l *log.Logger) {
	if l == nil {
		return
	}

	l.SetFormatter(log.StandardLogger().Formatter)
	l.SetOutput(log.StandardLogger().Out)
	l.SetLevel(log.StandardLogger().GetLevel())
}

// BracketFormatter writes logrus TTY-style lines, such as
// "INFO[0306] message  key=value", without color codes, for log files.
type BracketFormatter struct{}

// Format implements logrus.Formatter.
func (f *BracketFormatter) Format(entry *log.Entry) ([]byte, error) {
	level := strings.ToUpper(entry.Level.String())

	if len(level) > 4 {
		level = level[:4]
	}

	elapsed := int(entry.Time.Sub(processStart) / time.Second)
	message := strings.TrimSuffix(entry.Message, "\n")

	keys := make([]string, 0, len(entry.Data))

	for k := range entry.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var b bytes.Buffer

	fmt.Fprintf(&b, "%s[%04d] %-44s", level, elapsed, message)

	for _, k := range keys {
		fmt.Fprintf(&b, " %s=", k)
		appendValue(&b, entry.Data[k])
	}

	b.WriteByte('\n')

	return b.Bytes(), nil
}

func appendValue(b *bytes.Buffer, value any) {
	text, isString := value.(string)

	if !isString {
		fmt.Fprint(b, value)
		return
	}

	if needsQuoting(text) {
		b.WriteByte('"')
		b.WriteString(text)
		b.WriteByte('"')
		return
	}

	b.WriteString(text)
}

func needsQuoting(text string) bool {
	if text == "" {
		return true
	}

	for _, ch := range text {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '-' && ch != '.' && ch != '_' {
			return true
		}
	}

	return false
}

type fileHook struct {
	writer    io.Writer
	formatter log.Formatter
}

func (h *fileHook) Levels() []log.Level {
	return log.AllLevels
}

func (h *fileHook) Fire(entry *log.Entry) error {
	line, err := h.formatter.Format(entry)

	if err != nil {
		return err
	}

	_, err = h.writer.Write(line)

	return err
}

// AddFileHook copies every entry from the standard logger to the file at path,
// using formatter. The file is truncated, and its directory created if needed.
func AddFileHook(path string, formatter log.Formatter) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create log directory %s: %w", dir, err)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)

	if err != nil {
		return fmt.Errorf("open log file %s: %w", path, err)
	}

	log.AddHook(&fileHook{writer: file, formatter: formatter})

	return nil
}
