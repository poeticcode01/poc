package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	redis_leaderboard "github.com/poeticcode01/poc/redis_leaderboard"
)

var lb *redis_leaderboard.Leaderboard

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	lb = redis_leaderboard.NewLeaderboard(rdb, "leaderboard:scores")

	mux := http.NewServeMux()
	mux.HandleFunc("/score", handleSubmitScore)
	mux.HandleFunc("/score/incr", handleIncrementScore)
	mux.HandleFunc("/top", handleTopN)
	mux.HandleFunc("/rank/", handleGetRank)
	mux.HandleFunc("/around/", handleGetAround)
	mux.HandleFunc("/count", handleCount)
	mux.HandleFunc("/remove/", handleRemove)

	log.Println("Leaderboard POC started, listening on :8080")
	if err := http.ListenAndServe(":8080", logRequest(mux)); err != nil {
		log.Fatal(err)
	}
}

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func handleSubmitScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	member := r.URL.Query().Get("member")
	scoreStr := r.URL.Query().Get("score")
	if member == "" || scoreStr == "" {
		http.Error(w, "member and score required", http.StatusBadRequest)
		return
	}
	score, err := redis_leaderboard.ParseScore(scoreStr)
	if err != nil {
		http.Error(w, "invalid score", http.StatusBadRequest)
		return
	}
	rank, err := lb.SubmitScore(r.Context(), member, score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": member, "score": score, "rank": rank})
}

func handleIncrementScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	member := r.URL.Query().Get("member")
	deltaStr := r.URL.Query().Get("delta")
	if member == "" || deltaStr == "" {
		http.Error(w, "member and delta required", http.StatusBadRequest)
		return
	}
	delta, err := redis_leaderboard.ParseScore(deltaStr)
	if err != nil {
		http.Error(w, "invalid delta", http.StatusBadRequest)
		return
	}
	newScore, err := lb.IncrementScore(r.Context(), member, delta)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": member, "score": newScore})
}

func handleTopN(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	n := int64(10)
	if s := r.URL.Query().Get("n"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
			n = v
		}
	}
	entries, err := lb.TopN(r.Context(), n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"top": entries})
}

func handleGetRank(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	member := r.URL.Path[len("/rank/"):]
	if member == "" {
		http.Error(w, "member required", http.StatusBadRequest)
		return
	}
	rank, score, err := lb.GetRank(r.Context(), member)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"member": member, "rank": rank, "score": score})
}

func handleGetAround(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	member := r.URL.Path[len("/around/"):]
	if member == "" {
		http.Error(w, "member required", http.StatusBadRequest)
		return
	}
	window := int64(5)
	if s := r.URL.Query().Get("window"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v >= 0 {
			window = v
		}
	}
	entries, err := lb.GetAround(r.Context(), member, window)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"around": entries})
}

func handleCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	count, err := lb.TotalCount(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": count})
}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	member := r.URL.Path[len("/remove/"):]
	if member == "" {
		http.Error(w, "member required", http.StatusBadRequest)
		return
	}
	if err := lb.Remove(r.Context(), member); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"removed": member})
}
