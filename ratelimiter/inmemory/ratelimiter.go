package inmemory

import (
	"fmt"
	"time"
)

type Algorithm string

const (
	AlgoTokenBucket Algorithm = "token_bucket"
)

type RateLimiter interface {
	Tokens() <-chan struct{}
}

type tokenBucketLimiter struct {
	tokens chan struct{}
}

func (t *tokenBucketLimiter) Tokens() <-chan struct{} {
	return t.tokens
}

func newTokenBucketLimiter(rate, burst int) *tokenBucketLimiter {
	tokens := make(chan struct{}, burst)

	for i := 0; i < burst; i++ {
		tokens <- struct{}{}
	}

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(rate))
		defer ticker.Stop()

		for range ticker.C {
			select {
			case tokens <- struct{}{}:
			default:
			}
		}
	}()

	return &tokenBucketLimiter{
		tokens: tokens,
	}
}

func NewRateLimiter(algo Algorithm, rate, burst int) (RateLimiter, error) {
	switch algo {
	case AlgoTokenBucket:
		return newTokenBucketLimiter(rate, burst), nil
	default:
		return nil, fmt.Errorf("unsupported rate limiting algorithm %q", algo)
	}
}
