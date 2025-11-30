// Edge server entry point
package main

import (
	"cdn-edge-server/internal/cache"
	"cdn-edge-server/internal/config"
	"cdn-edge-server/internal/server"
	// "cdn-edge-server/internal/ui"
)

func main() {
	// logic to run terminal CLI
	//ui.Run()

	// Initialize edge server's cache (load existing files if any)
	cache.Init()

	// Start TCP server and serve clients
	srv := server.NewTCPServer(config.EdgeHost, config.EdgePort, server.HandleClient)
	srv.ListenAndServe()
}
