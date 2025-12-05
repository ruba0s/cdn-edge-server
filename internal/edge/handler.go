// Edge server handler
package edge

import (
	"bufio"
	"cdn-edge-server/internal/cache"
	"cdn-edge-server/internal/config"
	"cdn-edge-server/internal/http"
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

	// Parse client request
	req, err := http.ParseReq(reader)

	if err != nil || req == nil {
		if err != nil && err.Error() == "EOF" {
			// Health check, silently ignore
			return
		}
		resp := http.BuildErrorResponse(400)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	handleRequest(conn, req)
}

func handleRequest(conn net.Conn, req *http.Request) {
	switch req.Method {
	case "GET":
		handleGET(conn, req.Path)
	case "HEAD":
		handleHEAD(conn, req.Path)
	case "POST", "PUT":
		handleWriteReq(conn, req)
	default:
		// Unsupported method
		resp := http.BuildErrorResponse(405)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
	}
}

// handleGet serves an HTTP GET request for the given path using the specified connection.
func handleGET(conn net.Conn, path string) {
	filename := filepath.Base(path)

	// Determine MIME (content-type header value)
	mimeType := getMimeType(filename)

	if cache.Has(filename) {
		// Cache hit
		dat, err := cache.Get(filename)
		if err != nil {
			// Edge server error (failed to load cache file)
			resp := http.BuildErrorResponse(500)
			conn.Write([]byte(resp.HeadString()))
			conn.Write(resp.Body)
			return
		}

		resp := http.BuildResponse(200, mimeType, dat)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	// Cache miss, fetch from origin
	originResp, err := fetchFromOrigin("GET", filename, nil)
	if err != nil {
		resp := http.BuildErrorResponse(502)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	// Cache file
	if originResp.Status == 200 {
		cache.Add(filename, originResp.Body)
	}

	// Forward origin server response to client
	conn.Write([]byte(originResp.HeadString()))
	conn.Write(originResp.Body)
}

// handleHead processes an HTTP HEAD request for the given path using the given connection.
func handleHEAD(conn net.Conn, path string) {
	filename := filepath.Base(path) // filename w/o path for local cache storage/lookup
	cachePath := filepath.Join(config.CacheDir, filename)

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
	originResp, err := fetchFromOrigin("HEAD", filename, nil)
	if err != nil {
		resp := http.BuildErrorResponse(502)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	// Forward origin server response to client (headers only)
	conn.Write([]byte(originResp.HeadString()))
}

// handleWriteReq proccesses a POST or PUT request for the given file path using the given connection
// by forwarding them to the origin server.
func handleWriteReq(conn net.Conn, req *http.Request) {
	filename := filepath.Base(req.Path) // filename w/o path for local cache storage/lookup

	// For POST/PUT requests, forward request to origin server
	originResp, err := fetchFromOrigin(req.Method, filename, req.Body)
	if err != nil {
		resp := http.BuildErrorResponse(502)
		conn.Write([]byte(resp.HeadString()))
		conn.Write(resp.Body)
		return
	}

	// Remove file from cache (for PUT requests) if write to origin succeeded
	if originResp.Status == 200 {
		cache.Remove(filename)
	}

	// Forward origin response to client
	conn.Write([]byte(originResp.HeadString()))
	conn.Write(originResp.Body)
}

// fetchFromOrigin forwards the client's HTTP request with the given method and filename to the origin server,
// and fetches and returns the origin server's response.
func fetchFromOrigin(method, filename string, body []byte) (*http.Response, error) {
	connOrigin, err := net.Dial("tcp", config.OriginHost+":"+config.OriginPort)
	if err != nil {
		return nil, err
	}
	defer connOrigin.Close()

	reqStr := fmt.Sprintf(
		"%s /%s HTTP/1.0\r\nHost: localhost\r\nContent-Length: %d\r\n\r\n",
		method, filename, len(body),
	)
	connOrigin.Write([]byte(reqStr))

	if len(body) > 0 {
		connOrigin.Write(body)
	}

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
