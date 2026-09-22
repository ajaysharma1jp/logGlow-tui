package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	cancel context.CancelFunc
}

func initialModel(c context.CancelFunc) model {
	return model{cancel: c}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":

			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "\n  logGlow-tui initializing...\n\n  Press 'q' or 'ctrl+c' to quit.\n"
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
	m := initialModel(cancel)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
