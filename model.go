package main

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd // capture commands from view port
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			// first time setup: create viewport (fit screen)
			m.vp = viewport.New(msg.Width, msg.Height-6)
			// restore those logs which arrived before screen was ready
			m.vp.SetContent(m.getVisibleLogs())
			m.vp.GotoBottom()
			m.ready = true
		} else {
			// for resize window, adjust viewport
			m.vp.Width = msg.Width
			m.vp.Height = msg.Height - 6
		}
	case logMsg:
		newEntry := parseLogLine(string(msg))
		m.logs = append(m.logs, newEntry)
		if len(m.logs) > 10000 {
			m.logs = m.logs[1:]
		}
		m.vp.SetContent(m.getVisibleLogs())
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

	if m.isSearching {
		s += "\n Search: " + m.searchInput.View() + "\n"
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" (Press Esc or Enter to exit search)")
	} else {
		s += " Press 'q' or 'ctrl+c' to quit.\n"
	}
	return s
}

func (m model) getVisibleLogs() string{
	var visible []string
	searchTerm := strings.ToLower(m.searchInput.Value())
	for _, entry := range m.logs{
		if(searchTerm==""||strings.Contains(strings.ToLower(entry.RawText),searchTerm)){
			visible = append(visible, styleLogLine(entry))
		}
	}
	return strings.Join(visible,"\n")
}