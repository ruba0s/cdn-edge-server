// Edge server handler
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
		fmt.Println("PANIC HERE LINE 28")
		panic(err)
	}

	switch req.Method {
	case "GET":
		handleGet(conn, req.Path)
	case "HEAD":
		handleHead(conn, req.Path)
	// case "POST":
	// 	handlePost(conn, req.Path)
	// case "PUT":
	// 	handlePut(conn, req.Path)
	default:
		fmt.Println("Use either GET or HEAD for now")
	}
}

// handleGet serves an HTTP GET request for the given path using the specified connection.
func handleGet(conn net.Conn, path string) {
	filename := filepath.Base(path)
	cachePath := "/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + filename

	// Determine MIME (content-type header value)
	mimeType := getMimeType(filename)

	dat, err := os.ReadFile(cachePath)
	if err == nil {
		// Cache hit
		resp := http.BuildResponse(200, mimeType, dat)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	// Cache miss, fetch from origin
	originResp, err := fetchFromOrigin("GET", filename)
	if err != nil {
		resp := http.BuildResponse(500, mimeType, nil)
		conn.Write([]byte(resp.HeadString()))
		return
	}

	// Cache file
	if len(originResp.Body) > 0 {
		os.WriteFile(cachePath, originResp.Body, 0644)
	}

	// Forward origin server response to client
	conn.Write([]byte(originResp.HeadString()))
	conn.Write(originResp.Body)
}

// handleHead processes an HTTP HEAD request for the given path using the given connection.
func handleHead(conn net.Conn, path string) {
	filename := filepath.Base(path) // filename w/o path for local cache storage/lookup
	cachePath := "/Users/Ruba/Dev/cdn-edge-server/internal/cache/" + filename

	// Determine MIME (content-type header value)
	mimeType := getMimeType(filename)

	// Cache hit (HEAD only checks file existence, does not read body)
	info, err := os.Stat(cachePath)
	if err == nil {
		// Build and send HEAD resp
		resp := http.BuildResponse(200, mimeType, nil).WithHeader("Content-Length", fmt.Sprint(info.Size()))

		conn.Write([]byte(resp.HeadString()))
		return
	}

	// Cache miss, forward HEAD request to origin
	originResp, err := fetchFromOrigin("HEAD", filename)
	if err != nil {
		fmt.Println("DEBUG: origin resp", originResp, "error", err)
		resp := http.BuildResponse(originResp.Status, mimeType, nil)
		conn.Write([]byte(resp.HeadString()))
		return
	}

	// Forward origin server response to client (headers only)
	conn.Write([]byte(originResp.HeadString()))
}

// fetchFromOrigin forwards the client's HTTP request with the given method and filename to the origin server,
// and fetches and returns the origin server's response.
func fetchFromOrigin(method string, filename string) (*http.Response, error) {
	connOrigin, err := net.Dial("tcp", ORIGIN_HOST+":"+ORIGIN_PORT)
	if err != nil {
		return nil, err
	}
	defer connOrigin.Close()

	req := fmt.Sprintf(method+" /%s HTTP/1.0\r\nHost: localhost\r\n\r\n", filename)
	connOrigin.Write([]byte(req))

	reader := bufio.NewReader(connOrigin)
	resp, err := http.ParseResp(reader)
	if resp == nil || (err != nil && method != "HEAD" && err.Error() != "EOF") {
		// EOF errors are ignored for HEAD requests (unless the origin server returns a nil resp)
		return nil, err
	}

	return resp, nil
}

// getMimeType returns the MIME tyope of the given file's name via its extension.
// (Default: arbitrary binary data)
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // fallback (arbitrary binary data)
	}

	return mimeType
}
