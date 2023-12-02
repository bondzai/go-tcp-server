package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Server represents the TCP server.
type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
}

// NewServer creates a new instance of the TCP server.
func NewServer(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &Server{
		listener: listener,
	}, nil
}

// Start starts the TCP server.
func (s *Server) Start() {
	fmt.Printf("Server is listening on %s\n", s.listener.Addr().String())

	// Handle shutdown gracefully
	go s.handleSignals()

	for {
		conn, err := s.listener.Accept()

		// Check if the error is due to an interrupt signal
		if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
			fmt.Println("Server has been interrupted. Shutting down...")
			break
		}

		if err != nil {
			fmt.Printf("Error accepting connection: %s\n", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles individual client connections.
func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	// Echo back received data
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading data: %s\n", err)
			return
		}

		data := buffer[:n]
		fmt.Printf("Received: %s", data)

		_, err = conn.Write(data)
		if err != nil {
			fmt.Printf("Error writing data: %s\n", err)
			return
		}
	}
}

// handleSignals listens for interrupt signals to gracefully shutdown the server.
func (s *Server) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh

	fmt.Println("\nShutting down gracefully...")
	s.listener.Close()
	s.wg.Wait()

	os.Exit(0)
}

func main() {
	server, err := NewServer("127.0.0.1:8080")
	if err != nil {
		fmt.Printf("Error creating server: %s\n", err)
		os.Exit(1)
	}

	server.Start()
}
