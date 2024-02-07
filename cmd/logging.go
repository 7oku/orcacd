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

var logOrcacd = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	Prefix:          lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render("orcacd"),
	ReportTimestamp: true,
	Level:           LogLevelFromString(config.Loglevel),
})

var logCompose = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	Prefix:          lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Render("compose"),
	ReportTimestamp: true,
	Level:           LogLevelFromString(config.Loglevel),
})

var logPuller = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	Prefix:          lipgloss.NewStyle().Foreground(lipgloss.Color("#0f00ff")).Render("puller"),
	ReportTimestamp: true,
	Level:           LogLevelFromString(config.Loglevel),
})

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
