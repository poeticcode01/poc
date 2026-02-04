package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func main() {
	// Database connection string from environment variable
	dbConnectionString := os.Getenv("DATABASE_URL")
	if dbConnectionString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Initialize database connection
	err := InitDB(dbConnectionString)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create seats table and seed initial data
	err = CreateSeatsTable()
	if err != nil {
		log.Fatalf("Failed to create seats table: %v", err)
	}

	err = SeedSeats(100) // Seed 100 seats for testing
	if err != nil {
		log.Fatalf("Failed to seed seats: %v", err)
	}

	// Set up HTTP server
	http.HandleFunc("/reserve", reserveSeatHandler)
	http.HandleFunc("/status", seatStatusHandler) // New endpoint to get seat status
	http.HandleFunc("/available", availableSeatsHandler) // New endpoint to get available seats

	port := ":8080"
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// New handler for available seats
func availableSeatsHandler(w http.ResponseWriter, r *http.Request) {
	availableSeats, err := GetAvailableSeats()
	if err != nil {
		http.Error(w, "Failed to retrieve available seats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableSeats)
}

// New handler for seat status
func seatStatusHandler(w http.ResponseWriter, r *http.Request) {
	seats, err := GetAllSeats()
	if err != nil {
		http.Error(w, "Failed to retrieve seat status", http.StatusInternalServerError)
		return
	}

	DisplaySeatMatrix(w, seats)
}

func DisplaySeatMatrix(w http.ResponseWriter, seats []Seat) {
	// Assuming seat numbers are like A01, A02, ..., A10, B01, B02, ..., B10 etc.
	// We'll determine rows and columns dynamically.
	seatMap := make(map[string]Seat)
	maxRow := byte('A') // Initialize as byte
	maxCol := 0

	for _, seat := range seats {
		seatMap[seat.SeatNumber] = seat
		if len(seat.SeatNumber) > 0 {
			rowChar := seat.SeatNumber[0] // rowChar is byte
			if rowChar > maxRow {
				maxRow = rowChar
			}
			colNum, err := strconv.Atoi(seat.SeatNumber[1:])
			if err == nil && colNum > maxCol {
				maxCol = colNum
			}
		}
	}

	fmt.Fprintln(w, "\nSeat Reservation Status:")
	fmt.Fprint(w, "   ")
	for c := 1; c <= maxCol; c++ {
		fmt.Fprintf(w, " %02d", c)
	}
	fmt.Fprintln(w)

	for r := byte('A'); r <= maxRow; r++ { // Iterate using byte
		fmt.Fprintf(w, "%c: ", r)
		for c := 1; c <= maxCol; c++ {
			seatNumber := fmt.Sprintf("%c%02d", r, c)
			seat, ok := seatMap[seatNumber]
			if ok && seat.IsReserved {
				fmt.Fprint(w, " R ") // Reserved
			} else if ok && !seat.IsReserved {
				fmt.Fprint(w, " U ") // Unreserved
			} else {
				fmt.Fprint(w, "   ") // Non-existent seat or padding
			}
		}
		fmt.Fprintln(w)
	}
}

func reserveSeatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	seatNumber := r.URL.Query().Get("seat_number")
	if seatNumber == "" {
		http.Error(w, "seat_number parameter is required", http.StatusBadRequest)
		return
	}

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		log.Println("No client_id provided, using a default.")
		clientID = "anonymous_client"
	}

	reserved, err := ReserveSeat(seatNumber, clientID)
	if err != nil {
		log.Printf("Error reserving seat %s for client %s: %v\n", seatNumber, clientID, err)
		http.Error(w, fmt.Sprintf("Failed to reserve seat: %v", err), http.StatusInternalServerError)
		return
	}

	if reserved {
		log.Printf("Seat %s successfully reserved by client %s\n", seatNumber, clientID)
		fmt.Fprintf(w, "Seat %s successfully reserved by client %s\n", seatNumber, clientID)
	} else {
		log.Printf("Seat %s is already reserved\n", seatNumber)
		fmt.Fprintf(w, "Seat %s is already reserved\n", seatNumber)
	}
}
