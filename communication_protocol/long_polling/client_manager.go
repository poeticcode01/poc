package long_polling

import (
	"sync"
)

// ClientManager manages connected clients for long polling.
type ClientManager struct {
	clients map[string]chan string // Map clientID to a channel for updates
	mu      sync.RWMutex
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[string]chan string),
	}
}

func (cm *ClientManager) RegisterClient(clientID string) chan string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ch := make(chan string, 1) // Buffered channel to avoid deadlocks
	cm.clients[clientID] = ch
	return ch
}

func (cm *ClientManager) DeregisterClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if ch, ok := cm.clients[clientID]; ok {
		close(ch)
		delete(cm.clients, clientID)
	}
}

func (cm *ClientManager) BroadcastUpdate(update string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, ch := range cm.clients {
		select {
		case ch <- update:
			// Update sent successfully
		default:
			// Client channel is full, skip (or handle error/logging)
		}
	}
}
