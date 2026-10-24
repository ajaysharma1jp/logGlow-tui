package main

import (
	"encoding/json"
	"strings"
	"bytes"
)

type LogEntry struct {
	RawText string
	Level   string
}

func parseLogLine(line string) LogEntry {
	entry := LogEntry{RawText: line, Level: "unknown"}
	
	var partial struct{
		Level string `json:"level"`
	}

	if err := json.Unmarshal([]byte(line), &partial); err == nil && partial.Level!=""{
		entry.Level = strings.ToLower(partial.Level)
	}
	return entry
}

func formatJSON(raw string) string{
	var buf bytes.Buffer
	err := json.Indent(&buf, []byte(raw), "", " ")
	if err!=nil{
		return raw
	}
	return buf.String()
}

func styleLogLine(entry LogEntry) string {
	switch entry.Level {
	case "error", "fatal", "err":
		return errorStyle.Render(entry.RawText)
	case "warn", "warning":
		return warnStyle.Render(entry.RawText)
	case "info":
		return infoStyle.Render(entry.RawText)
	case "debug":
		return debugStyle.Render(entry.RawText)
	default:
		return entry.RawText
	}
}
