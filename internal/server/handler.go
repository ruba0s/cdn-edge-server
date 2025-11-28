package server

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	ORIGIN_HOST = "127.0.0.1"
	ORIGIN_PORT = "4396"
)

func HandleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	var headers []string
	//var requestLine string // change headers to not contain the request line ???

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected.")
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // reached end of headers
		}
		fmt.Println("line:", line)

		headers = append(headers, line)
	}

	if len(headers) == 0 {
		return
	}

	fmt.Println(headers)

	// Parse request line
	requestLine := headers[0]
	parts := strings.Split(requestLine, " ")
	fmt.Println("parts:", parts)

	if len(parts) < 3 { // part(s) of the request line are missing
		return
	}

	reqMethod := parts[0]
	reqPath := parts[1]

	switch reqMethod {
	case "GET":
		handleGet(conn, parts)
	case "HEAD":
		handleHead(conn, parts)
	default:
		fmt.Println("Use either GET or HEAD for now")
	}
	fmt.Println("Method:", reqMethod, "Path:", reqPath)
}

/*
Look up file or cached content
Return headers + body
*/
func handleGet(conn net.Conn, reqParts []string) {
	// Parse request line (mathod, path, filename)
	path := reqParts[1]
	filename := filepath.Base(path) // filename w/o path for local cache storage/lookup

	// Check cache (if cache hit just return the requested file)
	dat, err := os.ReadFile("/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + filename)
	if err == nil {
		// Cache hit
		fmt.Println("Cache hit:", filename)

		// Get file type (for response content type header field)
		ext := strings.ToLower(filepath.Ext(filename))
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream" // fallback
		}

		// Build headers
		headers :=
			"HTTP/1.0 200 OK\r\n" +
				"Content-Length: " + fmt.Sprint(len(dat)) + "\r\n" +
				"Content-Type: " + mimeType + "\r\n" +
				"\r\n"

		// Send headers
		conn.Write([]byte(headers))

		// Send body
		conn.Write(dat)
		return
	}

	// Cache miss, forward request to origin server
	connOrigin, err := net.Dial("tcp", ORIGIN_HOST+":"+ORIGIN_PORT)
	if err != nil {
		fmt.Println("Error connecting to origin server:", err)
	}
	defer connOrigin.Close()

	// Send HTTP request to origin
	fmt.Println("DEBUG filename", filename)
	originReq := fmt.Sprintf("GET /%s HTTP/1.0\r\nHost: localhost\r\n\r\n", filename)
	connOrigin.Write([]byte(originReq))

	// Parse origin response
	originReader := bufio.NewReader(connOrigin)

	// Read origin headers
	var oHeaders []string
	for {
		line, err := originReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading origin headers:", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break // end headers
		}

		oHeaders = append(oHeaders, line)
	}

	fmt.Println("Origin headers:", oHeaders)

	// Extract content length, read body of origin http response
	var contentLength int
	for _, h := range oHeaders {
		if strings.HasPrefix(strings.ToLower(h), "content-length:") {
			fmt.Sscanf(h, "Content-Length: %d", &contentLength)
		}
	}

	body := make([]byte, contentLength)
	_, err = io.ReadFull(originReader, body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}

	// Cache the file
	cachePath := "/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + filename
	os.WriteFile(cachePath, body, 0644)

	// Forward origin response to clients (response + body)
	response := strings.Join(oHeaders, "\r\n") + "\r\n\r\n"
	conn.Write([]byte(response))
	conn.Write(body)
}

/*
Return only head...
*/
func handleHead(conn net.Conn, reqParts []string) {

}
