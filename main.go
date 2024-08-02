package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	content, errChan := readTestFiles("./sip_messages")
	printMessages(content, errChan)
}

// readTestFiles reads all files in the given directory that start with "test" and sends their content to the output channel.
// If an error occurs, it sends the error to the error channel.
func readTestFiles(directory string) (<-chan string, <-chan error) {
	output := make(chan string)
	errChan := make(chan error)

	go func(output chan<- string, errChan chan<- error) {
		defer close(output)
		defer close(errChan)

		files, err := os.ReadDir(directory)
		if err != nil {
			errChan <- err
			return
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), "test") {
				content, err := os.ReadFile(directory + "/" + file.Name())
				if err != nil {
					errChan <- err
					return
				}
				output <- fmt.Sprintf("Content of %s:\n%s\n", file.Name(), content)
			}
		}
	}(output, errChan)

	return output, errChan
}

// printMessages reads from the content and error channels and prints the messages to the console.
// It stops when both channels are closed.
func printMessages(content <-chan string, errChan <-chan error) {
	for {
		select {
		case c, ok := <-content:
			if !ok {
				content = nil
			} else {
				fmt.Println(c)
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				fmt.Println(err)
			}
		}

		if content == nil && errChan == nil {
			break
		}
	}
}