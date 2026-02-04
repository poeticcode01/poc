package main

import (
	"database/sql"
	"fmt"
)

// ReserveSeat attempts to reserve a seat, ensuring atomicity with transactions and SELECT FOR UPDATE.
func ReserveSeat(seatNumber, clientID string) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Ensure rollback if an error occurs before commit

	var isReserved bool
	var reservedBy sql.NullString

	// SELECT FOR UPDATE locks the row to prevent race conditions during concurrent updates.
	query := "SELECT is_reserved, reserved_by FROM seats WHERE seat_number = $1 FOR UPDATE"
	err = tx.QueryRow(query, seatNumber).Scan(&isReserved, &reservedBy)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("seat %s not found", seatNumber)
	}
	if err != nil {
		return false, fmt.Errorf("failed to query seat %s: %w", seatNumber, err)
	}

	if isReserved {
		return false, nil // Seat already reserved by someone else
	}

	// If the seat is available, reserve it.
	updateQuery := "UPDATE seats SET is_reserved = TRUE, reserved_by = $1 WHERE seat_number = $2"
	_, err = tx.Exec(updateQuery, clientID, seatNumber)
	if err != nil {
		return false, fmt.Errorf("failed to update seat %s: %w", seatNumber, err)
	}

	return true, tx.Commit() // Commit the transaction to finalize the reservation
}
