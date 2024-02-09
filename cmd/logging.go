package main

import (
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type MaskedWriter struct {
	writer io.Writer
}

func NewMaskedWriter() *MaskedWriter {
	return &MaskedWriter{writer: os.Stdout}
}

var maskableParameters = []string{"jwt", "token", "password", "pass", "passphrase", "secret"}
var pattern = `([^&=?]+)=([^&]*)`
var regex = regexp.MustCompile(pattern)
var maskedValue = "<redacted>"

func (m MaskedWriter) Write(p []byte) (int, error) {
	stringToLog := string(p[:])

	for _, key := range maskableParameters {
		stringToLog = regex.ReplaceAllStringFunc(stringToLog, func(match string) string {
			parts := strings.SplitN(match, "=", 2)
			if len(parts) == 2 && parts[0] == key {
				return key + "=" + maskedValue
			}
			return match
		})
	}
	return m.writer.Write([]byte(stringToLog + "\n"))
}

// new logger
func NewLogger(name string, color string, level string) *log.Logger {
	var logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		Prefix:          lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(name),
		ReportTimestamp: true,
		Level:           LogLevelFromString(level),
	})
	return logger
}

// extract loglevel
func LogLevelFromString(loglevel string) log.Level {
	loglevel = strings.ToLower(loglevel)

	switch loglevel {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.DebugLevel
	}
}

// instanciate loggers
var logOrcacd = NewLogger("orcacd", config.Loglevel, "#333333")
var logPuller = NewLogger("puller", config.Loglevel, "#0f00ff")
var logCompose = NewLogger("compose", config.Loglevel, "#ff0000")
