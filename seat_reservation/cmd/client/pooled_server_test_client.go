package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func simulateClients(numClients int) {
	log.Printf("Starting client simulation with %d clients.\n", numClients)

	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()
			clientName := fmt.Sprintf("client-%d", clientID)

			// 1. Fetch available seats
			availableSeats, err := fetchAvailableSeats()
			if err != nil {
				log.Printf("Client %s: Failed to fetch available seats: %v\n", clientName, err)
				return
			}

			if len(availableSeats) == 0 {
				log.Printf("Client %s: No seats available to book.\n", clientName)
				return
			}

			// 2. Randomly select a seat
			seatToBook := availableSeats[rand.Intn(len(availableSeats))]
			log.Printf("Client %s: Attempting to book seat %s\n", clientName, seatToBook)

			// 3. Attempt to book the selected seat
			reserveURL := fmt.Sprintf("http://localhost:8080/reserve?seat_number=%s&client_id=%s", seatToBook, clientName)
			resp, err := http.Post(reserveURL, "application/json", nil)
			if err != nil {
				log.Printf("Client %s: Error making reservation request for seat %s: %v\n", clientName, seatToBook, err)
				return
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Client %s: Error reading reservation response for seat %s: %v\n", clientName, seatToBook, err)
				return
			}

			log.Printf("Client %s: Reservation response for seat %s: %s - %s\n", clientName, seatToBook, resp.Status, string(body))

			time.Sleep(50 * time.Millisecond) // Simulate some delay between clients

		}(i)
	}

	wg.Wait()
	log.Println("Client simulation finished.")
}

func fetchAvailableSeats() ([]string, error) {
	resp, err := http.Get("http://localhost:8080/available")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to /available endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK status from /available: %s", resp.Status)
	}

	var availableSeats []string
	err = json.NewDecoder(resp.Body).Decode(&availableSeats)
	if err != nil {
		return nil, fmt.Errorf("failed to decode available seats: %w", err)
	}

	return availableSeats, nil
}

func main() {
	// The main server should be running in a separate process.

	log.Println("Starting client simulation from main...")
	simulateClients(100) // Simulate 10 clients attempting to book seats

	// After simulation, fetch and display seat status
	log.Println("Fetching final seat status...")
	resp, err := http.Get("http://localhost:8080/status")
	if err != nil {
		log.Fatalf("Failed to get seat status: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read status response body: %v", err)
	}

	fmt.Println(string(body))
}
