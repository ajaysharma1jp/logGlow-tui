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

func checkInfoPipe() error {
	info, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("Error getting Stdin info: %w\n", err)
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

	// a goroutine(background worker for ingest)
	go func() {
		if err := ingest(os.Stdin, linesChan); err != nil {
			errorsChan <- err
		}
	}()

	// main thread('range' conti. read from channgel until close(outChar)' called
	for line := range linesChan {
		fmt.Println(line)
	}

	// check for background worker reported error before closing
	select {
	case err := <-errorsChan:
		fmt.Fprintf(os.Stderr, "Background ingest failed: %v\n", err)
		os.Exit(1)
	default:
		// no error in channel exit cleanly
	}

}
