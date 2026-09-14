package main

import (
	"bufio"
	"fmt"
	"os"
)

// file responsible for only ingestion

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

func ingest() error {
	if err := checkInfoPipe(); err != nil {
		return err
	}

	fmt.Println("Ingesting...")
	// make a scanner obj of bufio
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Printf("Ingested: %s\n", text)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error in reading standard input: %w\n", err)
	}

	return nil
}
