package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
	cancel       context.CancelFunc
	linesChan    chan string // channel for logs
	logs         []LogEntry  // list for logs
	filteredLogs []LogEntry
	ready        bool
	vp           viewport.Model
	isSearching  bool
	searchInput  textinput.Model
}

type LogEntry struct {
	RawText string
	Level   string
}

func initialModel(c context.CancelFunc, ch chan string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter logs"
	ti.CharLimit = 156
	ti.Width = 40
	return model{
		cancel:      c,
		linesChan:   ch,
		logs:        []LogEntry{},
		searchInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return waitForLog(m.linesChan)
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

func parseLogLine(line string) LogEntry{
	entry := LogEntry{RawText: line, Level: "unknown"}
	return entry;
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd // capture commands from view port
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// first time setup: create viewport (fit screen)
			m.vp = viewport.New(msg.Width, msg.Height-6)
			// restore those logs which arrived before screen was ready
			m.vp.SetContent(strings.Join(m.logs, "\n"))
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
		if len(m.logs) > 10000 {
			m.logs = m.logs[1:]
		}
		m.vp.SetContent(strings.Join(m.logs, "\n"))
		m.vp.GotoBottom() // auto scroll bottom to see newest log
		return m, waitForLog(m.linesChan)
	case tea.KeyMsg:
		if m.isSearching {
			switch msg.String() {
			case "enter", "esc":
				m.isSearching = false
				m.searchInput.Blur()
			default:
				m.searchInput, cmd = m.searchInput.Update(msg) // pass all other keystroke to text input bubble
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				if m.cancel != nil {
					m.cancel()
				}
				return m, tea.Quit
			case "/":
				m.isSearching = true
				m.searchInput.Focus()
				return m, nil // does not type '/' in searchbox while toggling it
			}
		}
	}

	// for message wasn't 'q' or log(like arrow keys or mouse scrolls) give it to vp for handling scrolling
	m.vp, cmd = m.vp.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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
	
	if m.isSearching{
		s+="\n Search: "+m.searchInput.View()+"\n"
		s+=lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" (Press Esc or Enter to exit search)")
	}else{
		s += " Press 'q' or 'ctrl+c' to quit.\n"
	}
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
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
