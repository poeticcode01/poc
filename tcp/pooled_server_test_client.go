package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	serverAddr          = "localhost:8081"
	numConnections      = 20                    // Number of connections to attempt
	concurrentClients   = 10                    // How many client goroutines to launch in parallel
	initialConnectDelay = 50 * time.Millisecond // Delay between launching client goroutines
)

func main() {
	log.Println("--- Pooled Server Test Client Started ---")
	log.Printf("Connecting to %s with %d concurrent clients, sending %d requests.\n", serverAddr, concurrentClients, numConnections)

	var wg sync.WaitGroup
	results := make(chan string, numConnections) // Channel to collect results

	// Use a semaphore to limit concurrent client goroutines
	sem := make(chan struct{}, concurrentClients)

	for i := 0; i < numConnections; i++ {
		sem <- struct{}{} // Acquire a token, block if `concurrentClients` are already running
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }() // Release the token when client goroutine finishes

			conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second) // 1-second dial timeout
			if err != nil {
				results <- fmt.Sprintf("Client %d: Connection failed: %v", id, err)
				return
			}
			defer conn.Close()

			request := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
			_, err = conn.Write([]byte(request))
			if err != nil {
				if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
					results <- fmt.Sprintf("Client %d: Rejected by Server (Capacity Exceeded)", id)
				} else {
					results <- fmt.Sprintf("Client %d: Write failed: %v", id, err)
				}
				return
			}

			response, err := ioutil.ReadAll(conn)
			if err != nil {
				if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "read: connection reset by peer") {
					results <- fmt.Sprintf("Client %d: Rejected by Server (Capacity Exceeded)", id)
				} else {
					results <- fmt.Sprintf("Client %d: Read failed: %v", id, err)
				}
				return
			}

			// For this simulation, we're expecting "OK" from the server
			responseStr := strings.TrimSpace(string(response))
			if strings.Contains(responseStr, "HTTP/1.1 200 OK") {
				results <- fmt.Sprintf("Client %d: Processed (OK)", id)
			} else if responseStr == "Hello from raw TCP worker pool server!" {
				results <- fmt.Sprintf("Client %d: Generic TCP response (Queue or default path)", id)
			} else {
				results <- fmt.Sprintf("Client %d: Unexpected response: %q", id, responseStr)
			}

		}(i + 1)
		time.Sleep(initialConnectDelay) // Small delay between launching clients to avoid overwhelming system calls
	}

	wg.Wait() // Wait for all client goroutines to complete
	close(results)

	// Print all collected results
	fmt.Println("\n--- Test Results ---")
	processedCount := 0
	rejectedCount := 0
	otherCount := 0

	for res := range results {
		fmt.Println(res)
		if strings.Contains(res, "Processed (OK)") {
			processedCount++
		} else if strings.Contains(res, "Rejected by Server (Capacity Exceeded)") {
			rejectedCount++
		} else {
			otherCount++
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Processed (OK): %d\n", processedCount)
	fmt.Printf("  Rejected: %d\n", rejectedCount)
	fmt.Printf("  Other/Errors: %d\n", otherCount)
	fmt.Println("--- Pooled Server Test Client Finished ---")
}
