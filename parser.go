package main

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	debugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
)

type LogEntry struct {
	RawText string
	Level   string
}

func parseLogLine(line string) LogEntry {
	entry := LogEntry{RawText: line, Level: "unknown"}
	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(line), &logData); err == nil {
		if levelVal, exists := logData["level"]; exists {
			if levelStr, ok := levelVal.(string); ok {
				entry.Level = strings.ToLower(levelStr)
			}
		}
	}
	return entry
}

func styleLogLine(entry LogEntry) string {
	switch entry.Level {
	case "error", "fatal", "err":
		return errorStyle.Render(entry.RawText)
	case "warn", "warning":
		return errorStyle.Render(entry.RawText)
	case "info":
		return infoStyle.Render(entry.RawText)
	case "debug":
		return debugStyle.Render(entry.RawText)
	default:
		return entry.RawText
	}
}
