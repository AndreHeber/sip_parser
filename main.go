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

	// Format of the start line
	// Request-Line  =  Method SP Request-URI SP SIP-Version CRLF
	// Method        =  INVITE / ACK / OPTIONS / CANCEL / BYE / REGISTER / REFER / MESSAGE / extension-method
	// SP			=  "[<>]"
	// Request-URI   =  SIP-URI / SIPS-URI / absoluteURI
	// SIP-Version   =  "SIP" "/" 1*DIGIT "." 1*DIGIT
	// CRLF			=  "\r\n"
	// SIP-URI       =  "sip:" [ userinfo "@" ] hostport
	// SIPS-URI      =  "sips:" [ userinfo "@" ] hostport
	// userinfo      =  ( unreserved / escaped / user / password ) *( ";" param )
	// user          =  *( unreserved / escaped / user-unreserved )
	// password      =  *( unreserved / escaped / password-unreserved )
	// hostport      =  host [ ":" port ]
	// host          =  hostname / IPv4address / IPv6reference
	// hostname      =  *( domainlabel "." ) toplabel [ "." ]
	// domainlabel   =  alphanum / alphanum *( alphanum / "-" ) alphanum
	// toplabel      =  ALPHA / ALPHA *( alphanum / "-" ) alphanum
	// IPv4address   =  1*digit "." 1*digit "." 1*digit "." 1*digit
	// IPv6reference =  "[" IPv6address "]"
	// IPv6address   =  hexpart [ ":" IPv4address ]
	// hexpart       =  hexseq / hexseq "::" [ hexseq ] / "::" [ hexseq ]
	// hexseq        =  hex4 *( ":" hex4)
	// hex4          =  1*4HEXDIG
	// port          =  1*DIGIT
	// param         =  pname [ "=" pvalue ]
	// pname         =  pname-value *( ";" pname-value )
	// pname-value   =  token / quoted-string
	// pvalue        =  token / quoted-string
	// token         =  1*alphanum
	// quoted-string =  DQUOTE *( qdtext / quoted-pair ) DQUOTE
	// qdtext        =  LWS / %x21 / %x23-5B / %x5D-7E / UTF8-NONASCII
	// quoted-pair   =  "\" (%x00-09 / %x0B-0C / %x0E-7F)
	// LWS           =  [*WSP CRLF] 1*WSP
	// WSP           =  SP / HTAB
	// Status-Line   =  SIP-Version SP Status-Code SP Reason-Phrase CRLF
	// Reason-Phrase =  *(reserved / unreserved / escaped / UTF8-NONASCII / UTF8-CONT / SP / HTAB)
	// Status-Code   =  Informational / Redirection / Success / Client-Error / Server-Error
	// Informational =  1*DIGIT
	// Redirection   =  3DIGIT
	// Success       =  2DIGIT
	// Client-Error  =  4DIGIT
	// Server-Error  =  5DIGIT
	// unreserved    =  alphanum / mark
	// mark          =  "-" / "_" / "." / "!" / "~" / "*" / "'" / "(" / ")"
	// escaped       =  "%" HEXDIG HEXDIG
	// alphanum      =  ALPHA / DIGIT
	// ALPHA         =  %x41-5A / %x61-7A
	// DIGIT         =  %x30-39
	// HEXDIG        =  DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
	// UTF8-NONASCII =  %x80-FF
	// UTF8-CONT     =  %x80-BF
	// HEXDIG        =  DIGIT / "A" / "B" / "C" / "D" / "E" / "F"
	// parse the start line
	parts := strings.Split(msg.StartLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid start line: %s, not 3 parts, linenumber: %d", msg.StartLine, lineNumber)
	}

	if parts[2] == "SIP/2.0" {
		return nil, fmt.Errorf("invalid start line: %s, not SIP/2.0, linenumber: %d", msg.StartLine, lineNumber)
	}

	if parts[0] == "INVITE" {
		fmt.Println("INVITE")
	} else if parts[0] == "ACK" {
		fmt.Println("ACK")
	} else if parts[0] == "OPTIONS" {
		fmt.Println("OPTIONS")
	} else if parts[0] == "CANCEL" {
		fmt.Println("CANCEL")
	} else if parts[0] == "BYE" {
		fmt.Println("BYE")
	} else if parts[0] == "REGISTER" {
		fmt.Println("REGISTER")
	} else if parts[0] == "REFER" {
		fmt.Println("REFER")
	} else if parts[0] == "MESSAGE" {
		fmt.Println("MESSAGE")
	} else {
		fmt.Println("extension-method")
	}

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
