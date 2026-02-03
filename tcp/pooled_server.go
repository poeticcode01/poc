package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/poeticcode01/poc/tcp/workerpool"
)

// handleConnection processes an individual TCP connection. It simulates work and responds.
func handleConnection(conn net.Conn) {
	defer conn.Close() // Ensure the connection is closed when the function exits

	// Simulate work being done for 2 seconds
	time.Sleep(2 * time.Second)

	// Read incoming data from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Error reading from %s: %v\n", conn.RemoteAddr(), err)
		return
	}

	request := string(buffer[:n])
	fmt.Printf("Received from %s: %s\n", conn.RemoteAddr(), request)
	time.Sleep(5 * time.Second)
	// Check if the received data is an HTTP GET request for the root path
	if strings.HasPrefix(request, "GET / HTTP/1.1") {
		// Construct a valid HTTP 200 OK response
		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 2\r\n" +
			"\r\n" +
			"OK"
		conn.Write([]byte(response))
		fmt.Printf("Sent HTTP OK response to %s.\n", conn.RemoteAddr())
	} else {
		// For non-HTTP or other requests, send a generic TCP response
		conn.Write([]byte("Hello from raw TCP worker pool server!\n"))
		fmt.Printf("Sent generic TCP response to %s.\n", conn.RemoteAddr())
	}
}

func main() {
	// Setup OS signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Configure and create the worker pool
	maxWorkers := 3 // Limit to 3 concurrent connections
	pool := workerpool.NewWorkerPool(maxWorkers, 0, handleConnection)

	// Start listening for incoming TCP connections on port 8081
	listener, err := net.Listen("tcp", ":8081") // Using port 8081 to avoid conflict with main.go
	if err != nil {
		log.Fatalf("Error listening on port 8081: %v", err)
	}
	defer listener.Close()
	log.Printf("Pooled TCP Server listening on :8081 with %d workers", maxWorkers)

	// Goroutine to accept connections and submit them to the worker pool
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-stop:
					// Listener was closed due to an OS signal, exit gracefully
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue // Continue accepting other connections
				}
			}
			log.Printf("Accepted connection from %s, submitting to pool.", conn.RemoteAddr())
			pool.Submit(conn) // Submit the accepted connection to the worker pool
		}
	}()

	// Block main goroutine until an OS signal is received
	<-stop
	log.Println("Shutting down pooled TCP server...")

	// Initiate graceful shutdown for the listener and worker pool
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Close the listener (prevents new connections)
	if err := listener.Close(); err != nil {
		log.Printf("Error closing listener: %v", err)
	}

	// Stop the worker pool and wait for all active jobs to complete
	pool.Stop()

	log.Println("Pooled TCP server gracefully stopped")
}
