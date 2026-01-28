package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/poeticcode01/poc/ratelimiter/inmemory"
)

var limiter chan struct{}

func handler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-limiter:
		w.Write([]byte("OK"))
	default:
		w.WriteHeader(http.StatusTooManyRequests)
	}
}

func main() {
	limiter = inmemory.NewRateLimiter(5, 10)

	srv := &http.Server{
		Addr:    ":8085",
		Handler: http.HandlerFunc(handler),
	}

	// Goroutine to listen for OS signals for graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed:%+v", err)
		}
	}()

	log.Println("Server started on :8085")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %+v", err)
	}

	log.Println("Server gracefully stopped")
}