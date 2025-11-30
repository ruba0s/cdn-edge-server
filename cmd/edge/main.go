// Edge server entry point
package main

import (
	"cdn-edge-server/internal/cache"
	"cdn-edge-server/internal/server"
)

const (
	HOST = "127.0.0.1"
	PORT = "4390"
)

func main() {
	// logic to run terminal CLI (or GUI?)

	// Initialize edge server's cache (load existing files if any)
	cache.Init()

	// Start TCP server and serve clients
	srv := server.NewTCPServer(HOST, PORT, server.HandleClient)
	srv.ListenAndServe()
}
