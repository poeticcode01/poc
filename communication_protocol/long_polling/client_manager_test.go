package long_polling

import (
	"testing"
	"time"
)

// Test that registering and deregistering a client does not cause panics,
// and that broadcasting after deregistration does not try to send to a closed channel.
func TestRegisterAndDeregisterClient(t *testing.T) {
	cm := NewClientManager()

	ch := cm.RegisterClient("client-1")
	if ch == nil {
		t.Fatalf("expected non-nil channel from RegisterClient")
	}

	// Deregister should close and remove the channel from the manager.
	// If it were still in the map, BroadcastUpdate would panic when sending to a closed channel.
	cm.DeregisterClient("client-1")

	// This should not panic.
	cm.BroadcastUpdate("test-update")
}

// Test that BroadcastUpdate delivers messages to all currently registered clients.
func TestBroadcastUpdateToAllClients(t *testing.T) {
	cm := NewClientManager()

	c1 := cm.RegisterClient("c1")
	c2 := cm.RegisterClient("c2")

	msg := "hello"
	cm.BroadcastUpdate(msg)

	assertReceived := func(name string, ch chan string) {
		t.Helper()
		select {
		case got := <-ch:
			if got != msg {
				t.Fatalf("%s: expected %q, got %q", name, msg, got)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("%s: did not receive broadcasted message", name)
		}
	}

	assertReceived("c1", c1)
	assertReceived("c2", c2)
}

