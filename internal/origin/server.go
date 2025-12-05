package origin

import (
	"bufio"
	"cdn-edge-server/internal/config"
	"cdn-edge-server/internal/http"
	"fmt"
	"mime"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// StartOrigin starts the origin server.
func StartOrigin() {
	ln, _ := net.Listen("tcp", config.OriginHost+":"+config.OriginPort)
	fmt.Printf("Origin server running on %s ...\n", config.OriginHost+":"+config.OriginPort)

	for {
		conn, _ := ln.Accept()
		go handle(conn) // multithreaded origin
	}
}

func handle(conn net.Conn) {
	defer conn.Close()

	req, err := http.ParseReq(bufio.NewReader(conn))
	if err != nil || req == nil {
		return
	}

	filename := filepath.Base(req.Path)

	switch req.Method {
	case "GET":
		serveGET(conn, filename)
	case "HEAD":
		serveHEAD(conn, filename)
	case "POST":
		handlePOST(conn, filename, req.Body)
	case "PUT":
		handlePUT(conn, filename, req.Body)
	default:
		resp := http.NewResponse(400)
		conn.Write([]byte(resp.HeadString()))
	}
}

//
// GET + HEAD
//

func serveGET(conn net.Conn, filename string) {
	path := filepath.Join(config.StorageDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		write404(conn)
		return
	}

	resp := http.BuildResponse(200, detectMime(filename), data)
	conn.Write([]byte(resp.HeadString()))
	conn.Write(resp.Body)
}

func serveHEAD(conn net.Conn, filename string) {
	path := filepath.Join(config.StorageDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		write404(conn)
		return
	}

	resp := http.BuildResponse(200, detectMime(filename), data)
	conn.Write([]byte(resp.HeadString()))
}

// handlePOST writes a new file to storage using the given filename and body.
// It returns an error response if the file already exists (POST is create only).
func handlePOST(conn net.Conn, filename string, body []byte) {
	path := filepath.Join(config.StorageDir, filename)

	// Reject if file already exists (POST = create)
	if _, err := os.Stat(path); err == nil {
		resp := http.NewResponse(400).WithHeader("Error", "File already exists")
		conn.Write([]byte(resp.HeadString()))
		return
	}

	err := os.WriteFile(path, body, 0644)
	if err != nil {
		write500(conn)
		return
	}

	resp := http.NewResponse(200).WithHeader("Created", filename)
	conn.Write([]byte(resp.HeadString()))
	conn.Write(resp.Body)
}

// handlePUT creates or overwrites a file with the provided filename and body.
// It always writes the file (PUT is create or replace).
func handlePUT(conn net.Conn, filename string, body []byte) {
	path := filepath.Join(config.StorageDir, filename)

	// PUT = create or overwrite
	err := os.WriteFile(path, body, 0644)
	if err != nil {
		write500(conn)
		return
	}

	resp := http.NewResponse(200).WithHeader("Updated", filename)
	conn.Write([]byte(resp.HeadString()))
}

// write404 Writes a 404 error to the given connection.
func write404(conn net.Conn) {
	resp := http.NewResponse(404)
	conn.Write([]byte(resp.HeadString()))
}

// write500 Writes a 500 error to the given connection.
func write500(conn net.Conn) {
	resp := http.NewResponse(500)
	conn.Write([]byte(resp.HeadString()))
}

// getMimeType returns the MIME tyope of the given file's name via its extension.
// (Default: arbitrary binary data)
func detectMime(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // fallback (arbitrary binary data)
	}

	return mimeType
}
