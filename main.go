package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "\n  logGlow-tui initializing...\n\n  Press 'q' or 'ctrl+c' to quit.\n"
}

func main() {
	if err := ingest(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
