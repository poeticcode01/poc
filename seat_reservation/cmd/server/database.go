package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) error {
	var err error
	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL!")
	return nil
}

// CreateSeatsTable creates the 'seats' table if it doesn't exist
func CreateSeatsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS seats (
		id SERIAL PRIMARY KEY,
		seat_number VARCHAR(10) NOT NULL UNIQUE,
		is_reserved BOOLEAN NOT NULL DEFAULT FALSE,
		reserved_by VARCHAR(255)
	);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create seats table: %w", err)
	}

	log.Println("Seats table created or already exists.")
	return nil
}

// GetSeatByNumber retrieves a seat by its seat_number
func GetSeatByNumber(seatNumber string) (*Seat, error) {
	var seat Seat
	query := "SELECT id, seat_number, is_reserved, reserved_by FROM seats WHERE seat_number = $1"
	err := db.QueryRow(query, seatNumber).Scan(&seat.ID, &seat.SeatNumber, &seat.IsReserved, &seat.ReservedBy)
	if err == sql.ErrNoRows {
		return nil, nil // Seat not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get seat by number: %w", err)
	}
	return &seat, nil
}

// SeedSeats inserts initial seat data if the table is empty
func SeedSeats(count int) error {
	var rowCount int
	err := db.QueryRow("SELECT COUNT(*) FROM seats").Scan(&rowCount)
	if err != nil {
		return fmt.Errorf("failed to count seats: %w", err)
	}

	if rowCount == 0 {
		log.Printf("Seeding %d seats...\n", count)
		stmt, err := db.Prepare("INSERT INTO seats (seat_number, is_reserved, reserved_by) VALUES ($1, FALSE, NULL)")
		if err != nil {
			return fmt.Errorf("failed to prepare statement for seeding: %w", err)
		}
		defer stmt.Close()

		for i := 1; i <= count; i++ {
			seatNumber := fmt.Sprintf("A%02d", i)
			_, err := stmt.Exec(seatNumber)
			if err != nil {
				return fmt.Errorf("failed to insert seat %s: %w", seatNumber, err)
			}
		}
		log.Println("Seats seeded successfully.")
	} else {
		log.Printf("%d seats already exist, skipping seeding.\n", rowCount)
	}
	return nil
}

// Seat represents a seat in the database
type Seat struct {
	ID         int
	SeatNumber string
	IsReserved bool
	ReservedBy sql.NullString
}

// GetAllSeats retrieves all seats from the database
func GetAllSeats() ([]Seat, error) {
	rows, err := db.Query("SELECT id, seat_number, is_reserved, reserved_by FROM seats ORDER BY seat_number")
	if err != nil {
		return nil, fmt.Errorf("failed to get all seats: %w", err)
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var seat Seat
		err := rows.Scan(&seat.ID, &seat.SeatNumber, &seat.IsReserved, &seat.ReservedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan seat row: %w", err)
		}
		seats = append(seats, seat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return seats, nil
}

// GetAvailableSeats retrieves all unreserved seat numbers from the database
func GetAvailableSeats() ([]string, error) {
	rows, err := db.Query("SELECT seat_number FROM seats WHERE is_reserved = FALSE ORDER BY seat_number")
	if err != nil {
		return nil, fmt.Errorf("failed to get available seats: %w", err)
	}
	defer rows.Close()

	var availableSeats []string
	for rows.Next() {
		var seatNumber string
		err := rows.Scan(&seatNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to scan available seat number: %w", err)
		}
		availableSeats = append(availableSeats, seatNumber)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return availableSeats, nil
}

