package main

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test that calling the /updates handler returns an update
// when BroadcastUpdate sends a message.
func TestLongPollingHandlerReceivesUpdate(t *testing.T) {
	req := httptest.NewRequest("GET", "/updates?clientID=test-client", nil)
	w := httptest.NewRecorder()

	// In a separate goroutine, simulate an update being broadcast
	// shortly after the handler starts waiting on the channel.
	go func() {
		time.Sleep(50 * time.Millisecond)
		clientManager.BroadcastUpdate("test-update")
	}()

	longPollingHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	got := strings.TrimSpace(string(body))
	if got != "Update: test-update" {
		t.Fatalf("expected %q, got %q", "Update: test-update", got)
	}
}

