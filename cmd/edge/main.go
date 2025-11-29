// Edge server entry point
package main

import (
	"cdn-edge-server/internal/server"
)

const (
	HOST = "127.0.0.1"
	PORT = "4390"
)

func main() {
	// logic to run terminal CLI (or GUI?)

	// Start TCP server and serve clients
	srv := server.NewTCPServer(HOST, PORT, server.HandleClient)
	srv.ListenAndServe()
}
