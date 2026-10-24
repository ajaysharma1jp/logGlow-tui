package main

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewState int const(
	StateList ViewState = iota
	StateDetail
)

type model struct {
	LogLevel
	cancel    context.CancelFunc
	linesChan chan string

	buffer *LogBuffer

	filteredLogs []LogEntry

	ready       bool
	vp          viewport.Model
	isSearching bool
	searchInput textinput.Model

	state       ViewState
	cursor		int
	detailVp	viewport.Model
}

func initialModel(c context.CancelFunc, ch chan string) model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter logs"
	ti.CharLimit = 156
	ti.Width = 40
	return model{
		LogLevel:    NewLogLevel(),
		cancel:      c,
		linesChan:   ch,
		buffer:      NewLogBuffer(10000),
		searchInput: ti,
		state:		 StateList,
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
		vpHeight := max(0, msg.Height-headerHeight-footerHeight-borderOffset)
		vpWidth := max(0, msg.Width-borderOffset)
		if !m.ready {
			m.vp = viewport.New(vpWidth, vpHeight)
			m.detailVp = viewport.New(vpWidth,vpHeight)
			// restore those logs which arrived before screen was ready
			m.vp.SetContent(m.getVisibleLogs())
			m.vp.GotoBottom()
			m.ready = true
		} else {
			m.vp.Width = vpWidth
			m.vp.Height = vpHeight
			m.detailVp.Width = vpWidth
			m.detailVp.Height = vpHeight
		}
	case logMsg:
		newEntry := parseLogLine(string(msg))
		m.buffer.Add(newEntry)
		if m.state==StateList {
			m.vp.SetContent(m.getVisibleLogs())
		}
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
			case "up","k":
				if m.state==StateList && m.cursor>0{
					m.cursor--
				}
			case "down","j":
				// need a way to check max buffer length here
				if m.state==StateList{
					m.cursor++
				}
			case "enter":
				if m.state == StateList{
					m.state = StateDetail
					// TODO: Get the log at m.cursor, format it with formatJSON(), 
        			// and send it to m.detailVp.SetContent()
				}
			case "esc":
				if m.state == StateList{
					m.state = StateList
				}
			case "ctrl+c", "q":
				if m.cancel != nil {
					m.cancel()
				}
				return m, tea.Quit
			case "/":
				m.isSearching = true
				m.searchInput.Focus()
				return m, nil // does not type '/' in searchbox while toggling it

			default:
				if m.handleKeyStroke(msg.String()) {
					m.vp.SetContent(m.getVisibleLogs())
					m.vp.GotoBottom()
					return m, nil
				}
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
	var currentView string
	if m.state == StateDetail {
		currentView = m.detailVp.View()
	}else{
		currentView = m.vp.View()
	}
	
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
	var builder strings.Builder
	searchTerm := strings.ToLower(m.searchInput.Value())
	for i := 0; i < m.buffer.Len(); i++ {
		entry := m.buffer.At(i)
		if !m.IsVisible(entry.Level) {
			continue
		}
		if searchTerm == "" || strings.Contains(strings.ToLower(entry.RawText), searchTerm) {
			// Write the styled string directly to the builder
			builder.WriteString(styleLogLine(entry))
			builder.WriteString("\n")
		}
	}
	return strings.TrimSuffix(builder.String(), "\n")
}
