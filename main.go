package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

func main() {
	content, errChan := readTestFiles("./sip_messages")
	parseSIPMessages(content, errChan)
	printMessages(content, errChan)
}

type FileReader struct {
	Content io.Reader
	Filename string
}

// readTestFiles reads all files in the given directory that start with "test" and sends their content to the output channel.
// If an error occurs, it sends the error to the error channel.
func readTestFiles(directory string) (<-chan FileReader, <-chan error) {
	output := make(chan FileReader)
	errChan := make(chan error)

	go func(output chan<- FileReader, errChan chan<- error) {
		defer close(output)
		defer close(errChan)

		files, err := os.ReadDir(directory)
		if err != nil {
			errChan <- err
			return
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), "test") {
				fileHandle, err := os.Open(directory + "/" + file.Name())
				if err != nil {
					errChan <- err
					return
				}
				output <- FileReader { Content: fileHandle, Filename: file.Name() }
			}
		}
	}(output, errChan)

	return output, errChan
}

// printMessages reads from the content and error channels and prints the messages to the console.
// It stops when both channels are closed.
func printMessages(content <-chan FileReader, errChan <-chan error) {
	for {
		select {
		case c, ok := <-content:
			if !ok {
				content = nil
			} else {
				data, err := io.ReadAll(c.Content)
				if err != nil {
					fmt.Println("Error reading content:", err)
				} else {
					fmt.Println(string(data))
				}
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

func parseSIPMessages(file <-chan FileReader, errChan <-chan error) {
	for {
		select {
		case c, ok := <-file:
			if !ok {
				file = nil
			} else {
				msg, err := ParseSIP(c.Content)
				if err != nil {
					fmt.Printf("Parsing_SIP, file: %s, error %s\n", c.Filename, err.Error())
				} else {
					fmt.Printf("SIP: %v\n\n\n", msg)
				}
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				fmt.Println(err)
			}
		}

		if file == nil && errChan == nil {
			break
		}
	}
}

var headerPool = sync.Pool {
	New: func() interface{} {
		return make(map[string][]string)
	},
}

type SIPMessage struct {
	StartLine string
	Headers   map[string][]string
	Body      []byte
}

func (m *SIPMessage) Reset() {
	m.StartLine = ""
	for k := range m.Headers {
		delete(m.Headers, k)
	}
	m.Body = m.Body[:0]
}

func ParseSIP(reader io.Reader) (*SIPMessage, error) {
	msg := &SIPMessage{
		Headers: headerPool.Get().(map[string][]string),
	}

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // Preallocate 64KB, max 1MB
	scanner.Split(sipSplit)
	lineNumber := 0

	// Parse start line
	if !scanner.Scan() {
		return nil, io.EOF
	}
	msg.StartLine = scanner.Text()
	lineNumber++
	// fmt.Println(msg.StartLine)

	// Parse headers
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if line == "" {
			break // Empty line indicates end of headers
		}

		if line[0] == ' ' || line[0] == '\t' {
			// Continuation of previous header
			// lastHeader := len(msg.Headers) - 1
			// msg.Headers[lastHeader] += " " + strings.TrimSpace(line)
		} else {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid header: %s, linenumber: %d", line, lineNumber)
			}
			key := strings.ToLower(strings.TrimSpace(parts[0])) // Intern common headers
			value := strings.TrimSpace(parts[1])
			msg.Headers[key] = append(msg.Headers[key], value)
		}
	}

	// Parse body
	var bodyBuilder bytes.Buffer
	for scanner.Scan() {
		bodyBuilder.Write(scanner.Bytes())
		bodyBuilder.WriteByte('\n')
	}
	msg.Body = bodyBuilder.Bytes()

	return msg, nil
}

func sipSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		if i > 0 && data[i-1] == '\r' {
			// We have a CRLF-terminated line.
			return i + 1, data[0:i-1], nil
		}
		// We have a full newline-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
