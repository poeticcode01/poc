package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close() // Close the connection when the handler finishes

	// Read incoming data
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	request := string(buffer[:n])
	fmt.Printf("Received: %s\n", request)
	time.Sleep(10 * time.Second)

	// Check if it's an HTTP GET request for the root path
	if strings.HasPrefix(request, "GET / HTTP/1.1") {
		// Construct HTTP response
		response := "HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 2\r\n" +
			"\r\n" +
			"OK"
		conn.Write([]byte(response))
		fmt.Println("Sent HTTP OK response.")
	} else {
		// Respond to non-HTTP or other HTTP requests
		conn.Write([]byte("Hello from raw TCP server!\n"))
		fmt.Println("Sent generic TCP response.")
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error listening:", err)
		return // Exit if listener fails
	}
	defer listener.Close()
	fmt.Println("Listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue // Continue listening for other connections
		}
		go handleConnection(conn) // Handle each connection concurrently
	}
}
