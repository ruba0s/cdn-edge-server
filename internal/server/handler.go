package server

import (
	"bufio"
	"cdn-edge-server/internal/http"
	"fmt"
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

	// Parse client request
	req, err := http.ParseReq(reader)

	if err != nil || req == nil {
		panic(err)
	}

	switch req.Method {
	case "GET":
		handleGet(conn, req.Path)
	case "HEAD":
		handleHead(conn, req.Path)
	default:
		fmt.Println("Use either GET or HEAD for now")
	}
}

type headResult struct {
	Resp     *http.Response
	Filename string
	CacheHit bool
}

// handleGet serves an HTTP GET request for the given path using the specified connection.
// It calls handleHead to obtain metadata and content, writes the body to the connection on a cache
// hit, and caches the file on a miss.
func handleGet(conn net.Conn, path string) {
	res, err := handleHead(conn, path)
	if err != nil {
		panic(err) // ?????!!! replace with returning error respones
	}
	if res.CacheHit {
		// cache hit, return requested file
		conn.Write(res.Resp.Body)
	} else {
		// Cache miss, cache the file
		cachePath := "/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + res.Filename
		os.WriteFile(cachePath, res.Resp.Body, 0644)
	}
}

// handleHead processes an HTTP HEAD request for the given path using the given connection.
// It returns a headResult containing metadata about the requested resource (HTTP response, name of requested file, whether it was a cache hit)
// without sending a response body. If an error occurs while reading from the connection,
// validating the path, or generating the response headers, the error is returned.
func handleHead(conn net.Conn, path string) (*headResult, error) {
	filename := filepath.Base(path) // filename w/o path for local cache storage/lookup

	// Check cache (if cache hit just return the requested file)
	body, err := os.ReadFile("/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + filename)
	if err == nil {
		// Cache hit
		fmt.Println("Cache hit:", filename)

		// Get file type (for response content type header field)
		ext := strings.ToLower(filepath.Ext(filename))
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream" // fallback
		}

		// Build and send headers
		headers :=
			"HTTP/1.0 200 OK\r\n" +
				"Content-Length: " + fmt.Sprint(len(body)) + "\r\n" +
				"Content-Type: " + mimeType + "\r\n" +
				"\r\n"

		conn.Write([]byte(headers))
		// call Build Response !!!
		resp := http.BuildResponse(200, mimeType, body)

		res := &headResult{
			Resp:     resp,
			Filename: filename,
			CacheHit: true,
		}
		return res, nil
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
	resp, _ := http.ParseResp(originReader)
	fmt.Println(resp)

	// Forward origin response to clients (response + body)
	conn.Write([]byte(resp.HeadString()))
	return &headResult{
		Resp:     resp,
		Filename: filename,
		CacheHit: true,
	}, err
}
