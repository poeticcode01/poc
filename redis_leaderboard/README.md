# Redis Sorted Set Leaderboard POC

A small leaderboard service backed by **Redis sorted sets** (`ZADD`, `ZREVRANGE`, `ZREVRANK`, `ZSCORE`, `ZINCRBY`, `ZREM`, `ZCARD`).

## Why sorted sets?

- **Score ordering**: Members are kept sorted by score; higher score = better rank.
- **O(log N)** add/update/rank/score operations.
- **No duplicates**: One entry per member; updating score replaces the previous one.

## Run

### 1. Start Redis

```bash
docker compose up -d
```

### 2. Run the server

```bash
go run ./cmd
```

Server listens on `:8080`.

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/score?member=<id>&score=<float>` | Set or update a player's score; returns new rank |
| `POST` | `/score/incr?member=<id>&delta=<float>` | Add `delta` to current score (e.g. +100 points) |
| `GET`  | `/top?n=10` | Top N players (default 10) |
| `GET`  | `/rank/<member>` | Get rank (1-based) and score for a member |
| `GET`  | `/around/<member>?window=5` | Entries around member's rank (±window) |
| `GET`  | `/count` | Total number of players |
| `DELETE` | `/remove/<member>` | Remove a player from the leaderboard |

## Example

```bash
# Add scores
curl -X POST "http://localhost:8080/score?member=alice&score=1500"
curl -X POST "http://localhost:8080/score?member=bob&score=1200"
curl -X POST "http://localhost:8080/score?member=carol&score=1800"

# Top 10
curl "http://localhost:8080/top?n=10"

# Alice's rank and score
curl "http://localhost:8080/rank/alice"

# Leaderboard around Carol (window 2)
curl "http://localhost:8080/around/carol?window=2"

# Increment Bob by 100
curl -X POST "http://localhost:8080/score/incr?member=bob&delta=100"
```

## Redis keys

- `leaderboard:scores` — sorted set (score → member). Higher score = higher rank.

## Code layout

- `leaderboard.go` — core logic (submit, top N, rank, around, increment, count, remove).
- `cmd/main.go` — HTTP server wiring the leaderboard to the API above.
