package main

import (
	"bufio"
	"fmt"
	"io"
)

func ingest(reader io.Reader, outputChan chan<- string) error {
	defer close(outputChan)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		text := scanner.Text()
		outputChan <- text
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error in reading standard input: %w", err)
	}
	return nil
}
