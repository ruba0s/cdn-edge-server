// Edge server entry point
package main

import (
	"cdn-edge-server/internal/server"
	"fmt"
	"net"
)

const (
	HOST = "127.0.0.1"
	PORT = "4390"
)

func main() {
	// logic to run terminal CLI (or GUI?)

	// Start socket listener
	listener, err := net.Listen("tcp", HOST+":"+PORT)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on", HOST+":"+PORT)

	// Listen for incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Error occurred while accepting client connection, skip
			fmt.Println("Accept error:", err)
			continue
		}

		fmt.Println("Client connected from:", conn.RemoteAddr())

		go server.HandleClient(conn) // concurrently handle clients
	}
}
