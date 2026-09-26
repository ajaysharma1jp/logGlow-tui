package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type logMsg string

func waitForLog(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return logMsg(line)
	}
}

func FormatLogLine(line string) string {
	var logData map[string]interface{}

	// attempt to unmarshal to see if its structured JSON
	err := json.Unmarshal([]byte(line), &logData)
	if err != nil {
		return line // retruns as plain text not json
	}

	// extract level
	if levelVal, exits := logData["level"]; exits {
		levelStr, ok := levelVal.(string)
		if ok {
			switch strings.ToLower(levelStr) {
			case "error", "fatel", "err":
				return errorStyle.Render(line)
			case "warn", "warning":
				return warnStyle.Render(line)
			case "info":
				return infoStyle.Render(line)
			case "debug":
				return debugStyle.Render(line)
			}
		}
	}
	return line
}

func checkInfoPipe() error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Error getting Stdin info: %w", err)
	}

	// for character device (terminal) vs pipe
	if info.Mode()&os.ModeCharDevice != 0 {
		return fmt.Errorf("No Data Piped!\nUsage: echo 'Sample' | go run ingest.go")
	}

	return nil
}

func main() {
	if err := checkInfoPipe(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}

	linesChan := make(chan string)
	errorsChan := make(chan error, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a goroutine(background worker for ingest)
	go func() {
		if err := ingest(ctx, os.Stdin, linesChan); err != nil {
			errorsChan <- err
		}
	}()

	// Bubble Tea UI controls the foreground loop; stdin ingest runs in the background.
	m := initialModel(cancel, linesChan)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
