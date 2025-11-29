package http

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
}

type Response struct {
	Version    string // assumed HTTP/1.0
	Status     int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// ParseReq reads an HTTP request from the given bufio.Reader and parses the request line
// and headers until it encounters a blank line.
// Returns a populated Request on success, nil if the request is empty or malformed,
// and an error if the reader encounters an I/O issue.
func ParseReq(reader *bufio.Reader) (*Request, error) {
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected.")
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // end of headers
		}

		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return nil, nil
	}

	// Parse request line
	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) < 3 { // part(s) of the request line are missing
		return nil, nil
	}

	method := parts[0]
	path := parts[1]
	headers := make(map[string]string)
	for _, h := range lines[1:] {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue // skip malformed headers
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}

	req := &Request{
		Method:  method,
		Path:    path,
		Headers: headers,
	}

	return req, nil
}

func ParseResp(reader *bufio.Reader) (*Response, error) {
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading response headers:", err)
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // end of headers
		}

		lines = append(lines, line)
	}

	parts := strings.Split(lines[0], " ")
	ver := parts[0]
	statCode, _ := strconv.Atoi(parts[1])
	statTxt := strings.Join(parts[2:], "")

	headers := make(map[string]string)
	for _, h := range lines[1:] {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			continue // skip malformed headers
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		headers[key] = value
	}
	fmt.Println(headers)

	// Extract content length
	var contentLength int
	for _, h := range lines {
		if strings.HasPrefix(strings.ToLower(h), "content-length:") {
			fmt.Sscanf(h, "Content-Length: %d", &contentLength)
		}
	}

	// TODO: If method isn't HEAD, read body of HTTP response ??
	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	resp := &Response{
		Version:    ver,
		Status:     statCode,
		StatusText: statTxt,
		Headers:    headers,
		Body:       body,
	}

	return resp, nil
}

// FIX -- PUT IN response.go??
func (r *Response) HeadString() string {
	var b strings.Builder

	// Status line
	b.WriteString(fmt.Sprintf("%s %d %s\r\n", r.Version, r.Status, r.StatusText))

	// Headers
	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// End of headers
	b.WriteString("\r\n")

	return b.String()
}
