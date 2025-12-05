// Edge server entry point
package main

import (
	"cdn-edge-server/internal/cache"
	"cdn-edge-server/internal/config"
	"cdn-edge-server/internal/edge"
)

func main() {
	// Initialize edge server's cache (load existing files if any)
	cache.Init()

	// Start TCP server and serve clients
	srv := edge.NewTCPServer(config.EdgeHost, config.EdgePort, edge.HandleClient)
	srv.ListenAndServe()
}
