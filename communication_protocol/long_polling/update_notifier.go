package long_polling

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // For PostgreSQL example
)

type UpdateNotifier struct {
	clientManager *ClientManager
	db            *sql.DB // Add database connection
	interval      time.Duration
	stopChan      chan struct{}
}

// NewUpdateNotifier creates a new UpdateNotifier with a database connection.
func NewUpdateNotifier(cm *ClientManager, db *sql.DB, interval time.Duration) *UpdateNotifier {
	return &UpdateNotifier{
		clientManager: cm,
		db:            db,
		interval:      interval,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the periodic database checking.
func (un *UpdateNotifier) Start() {
	ticker := time.NewTicker(un.interval)
	defer ticker.Stop()

	lastChecked := time.Now()

	for {
		select {
		case <-ticker.C:
			update, err := un.checkDatabaseForUpdates(lastChecked)
			if err != nil {
				log.Printf("Error checking database for updates: %v", err)
				continue
			}
			if update != "" {
				un.clientManager.BroadcastUpdate(update)
			}
			lastChecked = time.Now()
		case <-un.stopChan:
			return
		}
	}
}

// Stop halts the update notifier.
func (un *UpdateNotifier) Stop() {
	close(un.stopChan)
}

// checkDatabaseForUpdates is a placeholder for actual database checking logic.
// It should query your database for changes since the lastChecked time.
// For demonstration, it simulates a check.
func (un *UpdateNotifier) checkDatabaseForUpdates(lastChecked time.Time) (string, error) {
	// This expects a table like:
	//   CREATE TABLE updates (
	//     id SERIAL PRIMARY KEY,
	//     message TEXT NOT NULL,
	//     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	//   );

	rows, err := un.db.Query(
		`SELECT message, created_at
         FROM updates
         WHERE created_at > $1
         ORDER BY created_at ASC`,
		lastChecked,
	)
	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var latestMsg string
	var latestTime time.Time

	for rows.Next() {
		var msg string
		var createdAt time.Time
		if err := rows.Scan(&msg, &createdAt); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		latestMsg = msg
		latestTime = createdAt
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("row iteration error: %w", err)
	}

	if latestMsg == "" {
		return "", nil
	}

	return fmt.Sprintf("New data: %s (at %s)", latestMsg, latestTime.Format(time.RFC3339)), nil
}
