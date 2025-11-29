package server

import (
	"fmt"
	"net"
)

type TCPServer struct {
	Host    string
	Port    string
	Handler func(net.Conn)
}

func NewTCPServer(host, port string, handler func(net.Conn)) *TCPServer {
	return &TCPServer{
		Host:    host,
		Port:    port,
		Handler: handler,
	}
}

func (s *TCPServer) ListenAndServe() error {
	// Start socket listener
	addr := s.Host + ":" + s.Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}

	fmt.Println("Server listening on", addr)

	// Listen for incoming client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Error occurred while accepting client connection, skip that client and keep listening
			fmt.Println("accept error:", err)
			continue
		}
		fmt.Println("Client connected from:", conn.RemoteAddr())

		// Concurrently handle client connections
		go HandleClient(conn)
	}
}
