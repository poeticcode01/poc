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
	// TODO: Replace this with your actual database query.
	// Example: Query a table for rows updated after 'lastChecked'.

	// Example for PostgreSQL:
	// var count int
	// err := un.db.QueryRow("SELECT COUNT(*) FROM your_table WHERE updated_at > $1", lastChecked).Scan(&count)
	// if err != nil {
	//	 return "", fmt.Errorf("failed to query database: %w", err)
	// }
	// if count > 0 {
	//	 return fmt.Sprintf("New data available since %s", lastChecked.Format(time.RFC3339)), nil
	// }

	// For now, simulate an update.
	if time.Since(lastChecked) > un.interval && time.Now().Second()%10 == 0 {
		return fmt.Sprintf("Simulated data update at %s", time.Now().Format(time.RFC3339)), nil
	}
	return "", nil
}
