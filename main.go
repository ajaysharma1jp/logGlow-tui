package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	debugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
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

type model struct {
	cancel    context.CancelFunc
	linesChan chan string // channel for logs
	logs      []string    // list for logs
	ready     bool
	vp        viewport.Model
}

func initialModel(c context.CancelFunc, ch chan string) model {
	return model{
		cancel:    c,
		linesChan: ch,
		logs:      []string{},
	}
}

func (m model) Init() tea.Cmd {
	return waitForLog(m.linesChan)
}

func FormatLogLine(line string) string{
	var logData map[string]interface{}

	// attempt to unmarshal to see if its structured JSON
	err := json.Unmarshal([]byte(line), &logData)
	if(err != nil){
		return line // retuns as plain text not json
	}

	// extract level
	if levelVal, exits := logData["level"]; exits{
		levelStr,ok:=levelVal.(string)
		if ok{
			switch strings.ToLower(levelStr){
			case "error","fatel","err":
				return errorStyle.Render(line)
			case "warn","warning":
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd // capture commands from view port
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg: // buuble tea used for knowing terminal size
		if !m.ready {
			// first time setup: create viewport (fit screen)
			m.vp = viewport.New(msg.Width, msg.Height-6)
			// restore those logs which arrived before screen was ready
			m.vp.SetContent(strings.Join(m.logs,"\n"))
			m.vp.GotoBottom()
			m.ready = true
		} else {
			// for resize window, adjust viewport
			m.vp.Width = msg.Width
			m.vp.Height = msg.Height - 6
		}
	case logMsg:
		formattedLine := FormatLogLine(string(msg))
		m.logs = append(m.logs, formattedLine)
		if len(m.logs)>10000{
			m.logs = m.logs[1:] 
		}
		m.vp.SetContent(strings.Join(m.logs, "\n"))
		m.vp.GotoBottom() // auto scroll bottom to see newest log
		return m, waitForLog(m.linesChan)
	}

	// for message wasn't 'q' or log(like arrow keys or mouse scrolls) give it to vp for handling scrolling
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n Initializing...\n"
	}
	
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	header := headerStyle.Render("\n LogGlow-tui: Live Logs\n")
	
	s := header
	s += "--------------------\n"
	s += m.vp.View()

	s += "\n--------------------------------------------------------\n"
	s += " Press 'q' or 'ctrl+c' to quit.\n"
	return s
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
	p := tea.NewProgram(m,tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
