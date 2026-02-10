package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	long_polling "github.com/poeticcode01/poc/communication_protocol/long_polling"

	_ "github.com/lib/pq" // For PostgreSQL example, replace with your database driver
)

// clientManager is shared across all long polling handlers
var clientManager = long_polling.NewClientManager()

func longPollingHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientID")
	if clientID == "" {
		http.Error(w, "clientID is required", http.StatusBadRequest)
		return
	}

	clientChan := clientManager.RegisterClient(clientID)
	defer clientManager.DeregisterClient(clientID)

	timeout := time.After(30 * time.Second) // Adjust timeout as needed

	select {
	case update := <-clientChan:
		fmt.Fprintf(w, "Update: %s", update)
	case <-timeout:
		fmt.Fprint(w, "Timeout: No new updates")
	case <-r.Context().Done():
		log.Printf("Client %s disconnected", clientID)
	}
}

func main() {
	// Initialize database connection
	connStr := "user=postgres password=postgres dbname=long_polling sslmode=disable"
	db, err := sql.Open("postgres", connStr) // Replace with your database driver and connection string
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database!")

	// Start the update notifier
	notifier := long_polling.NewUpdateNotifier(clientManager, db, 5*time.Second) // Check every 5 seconds
	go notifier.Start()

	http.HandleFunc("/updates", longPollingHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
