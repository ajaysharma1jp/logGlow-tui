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
		headerHeight := 3
		footerHeight := 3
		borderOffset := 4
		vpHeight := msg.Height - headerHeight - footerHeight - borderOffset
		vpWidth := msg.Width - borderOffset
		if !m.ready {
			m.vp = viewport.New(vpWidth, vpHeight)
			// restore those logs which arrived before screen was ready
			m.vp.SetContent(m.getVisibleLogs())
			m.vp.GotoBottom()
			m.ready = true
		} else {
			m.vp.Width = vpWidth
			m.vp.Height = vpHeight
		}
	case logMsg:
		newEntry := parseLogLine(string(msg))
		m.logs = append(m.logs, newEntry)
		if len(m.logs) > 10000 {
			m.logs = m.logs[1:]
		}
		m.vp.SetContent(m.getVisibleLogs())
		m.vp.GotoBottom()
		return m, waitForLog(m.linesChan)
	case tea.KeyMsg:
		if m.isSearching {
			switch msg.String() {
			case "enter", "esc":
				m.isSearching = false
				m.searchInput.Blur()
			default:
				m.searchInput, cmd = m.searchInput.Update(msg) // pass all other keystroke to text input bubble
				m.vp.SetContent(m.getVisibleLogs())
				m.vp.GotoBottom()
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

	header := headerStyle.Render("\n LogGlow-tui: Live Logs\n")
	logsView := containerStyle.Width(m.vp.Width).Render(m.vp.View())

	var footer string

	if m.isSearching {
		search := searchBoxStyle.Render("🔍 " + m.searchInput.View())
		help := helpStyle.Render(" (Press Esc or Enter to exit search)")
		footer = "\n" + lipgloss.JoinHorizontal(lipgloss.Center, search, help)
	} else {
		footer = "\n" + helpStyle.Render(" Press '/' to search • 'q' to quit")
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, logsView, footer)
}

func (m model) getVisibleLogs() string {
	var visible []string
	searchTerm := strings.ToLower(m.searchInput.Value())
	for _, entry := range m.logs {
		if searchTerm == "" || strings.Contains(strings.ToLower(entry.RawText), searchTerm) {
			visible = append(visible, styleLogLine(entry))
		}
	}
	return strings.Join(visible, "\n")
}
