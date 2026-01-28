package inmemory

import "time"

func NewRateLimiter(rate, burst int) chan struct{} {
	tokens := make(chan struct{}, burst)

	// Fill bucket
	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	// Refill tokens
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}
	}()

	return tokens
}
