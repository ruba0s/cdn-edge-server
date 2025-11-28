package server

import (
	"bufio"
	"fmt"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
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
		// CACHE HIT
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

	// If cache miss, request origin server

	// Cache the requested file

	// Send HTTP response with requested file

}

/*
Return only head...
*/
func handleHead(conn net.Conn, reqParts []string) {

}
