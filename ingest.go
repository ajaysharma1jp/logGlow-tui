package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
)

func ingest(ctx context.Context, reader io.Reader, outputChan chan<- string) error {
	defer close(outputChan)
	done := make(chan struct{})
	defer close(done)

	scanner := bufio.NewScanner(reader)

	
	go func(){
		select{
		case <-ctx.Done():
			if closer, ok := reader.(io.Closer); ok{
				closer.Close()
			}
		case <-done:
			return
		}
	}()

	for scanner.Scan() {
		text := scanner.Text()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case outputChan <- text:
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error in reading standard input: %w", err)
	}
	return nil
}
