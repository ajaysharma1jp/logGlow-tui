package main

import (
	"bufio"
	"fmt"
	"os"
)

// file responsible for only ingestion

func checkInfoPipe() {
	// for file info
	info, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting Stdin info: %s\n", err)
		os.Exit(1)
	}

	// for character device (terminal) vs pipe
	if info.Mode()&os.ModeCharDevice != 0 {
		fmt.Println("No Data Piped!\nUsage: echo 'Sample' | go run ingest.go")
		os.Exit(1)
	}

}

func ingest() {
	checkInfoPipe()
	fmt.Println("Ingesting...")
	// make a scanner obj of bufio
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Printf("Ingested: %s\n", text)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error in reading standard input: ", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func main() {
	ingest()
}
