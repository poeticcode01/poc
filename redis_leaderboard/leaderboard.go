package redis_leaderboard

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const defaultLeaderboardKey = "leaderboard:scores"

// Leaderboard uses Redis sorted set (score → member) for ranking.
// Higher score = better rank. Same score: lexicographic order of member.
type Leaderboard struct {
	client *redis.Client
	key    string
}

// NewLeaderboard creates a leaderboard service backed by Redis sorted set.
func NewLeaderboard(client *redis.Client, key string) *Leaderboard {
	if key == "" {
		key = defaultLeaderboardKey
	}
	return &Leaderboard{client: client, key: key}
}

// Entry represents a single leaderboard entry (player + score).
type Entry struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
	Rank   int64   `json:"rank"` // 1-based rank (1 = top)
}

// SubmitScore sets or updates a member's score. Higher score wins.
// Returns the member's new rank (1-based) and score.
func (lb *Leaderboard) SubmitScore(ctx context.Context, member string, score float64) (rank int64, err error) {
	pipe := lb.client.Pipeline()
	pipe.ZAdd(ctx, lb.key, redis.Z{Score: score, Member: member})
	pipe.ZRevRank(ctx, lb.key, member)
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	// ZRevRank returns 0-based index; we want 1-based rank
	r, _ := cmds[1].(*redis.IntCmd).Result()
	return r + 1, nil
}

// TopN returns the top n entries (highest scores first), 1-based rank.
func (lb *Leaderboard) TopN(ctx context.Context, n int64) ([]Entry, error) {
	if n <= 0 {
		n = 10
	}
	stop := n - 1
	results, err := lb.client.ZRevRangeWithScores(ctx, lb.key, 0, stop).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, len(results))
	for i, z := range results {
		member, _ := z.Member.(string)
		entries[i] = Entry{
			Member: member,
			Score:  z.Score,
			Rank:   int64(i) + 1,
		}
	}
	return entries, nil
}

// GetRank returns 1-based rank and score for a member. 0 rank means not found.
func (lb *Leaderboard) GetRank(ctx context.Context, member string) (rank int64, score float64, err error) {
	pipe := lb.client.Pipeline()
	pipe.ZRevRank(ctx, lb.key, member)
	pipe.ZScore(ctx, lb.key, member)
	cmds, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}
	rankCmd := cmds[0].(*redis.IntCmd)
	scoreCmd := cmds[1].(*redis.FloatCmd)
	r, rErr := rankCmd.Result()
	s, sErr := scoreCmd.Result()
	if rErr == redis.Nil || sErr == redis.Nil {
		return 0, 0, nil
	}
	if rErr != nil {
		return 0, 0, rErr
	}
	if sErr != nil {
		return 0, 0, sErr
	}
	return r + 1, s, nil
}

// GetAround returns entries around a member's rank (e.g. "players near me").
// halfWindow is how many above and below; total returned is at most 2*halfWindow+1.
func (lb *Leaderboard) GetAround(ctx context.Context, member string, halfWindow int64) ([]Entry, error) {
	rank, _, err := lb.GetRank(ctx, member)
	if err != nil || rank <= 0 {
		return nil, err
	}
	start := rank - halfWindow - 1
	if start < 0 {
		start = 0
	}
	stop := rank + halfWindow - 1
	results, err := lb.client.ZRevRangeWithScores(ctx, lb.key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, len(results))
	for i, z := range results {
		m, _ := z.Member.(string)
		entries[i] = Entry{Member: m, Score: z.Score, Rank: start + int64(i) + 1}
	}
	return entries, nil
}

// IncrementScore adds delta to member's current score (useful for games).
// If member doesn't exist, they are treated as 0. Returns new score.
func (lb *Leaderboard) IncrementScore(ctx context.Context, member string, delta float64) (newScore float64, err error) {
	newScore, err = lb.client.ZIncrBy(ctx, lb.key, delta, member).Result()
	return newScore, err
}

// TotalCount returns the number of members in the leaderboard.
func (lb *Leaderboard) TotalCount(ctx context.Context) (int64, error) {
	return lb.client.ZCard(ctx, lb.key).Result()
}

// Remove removes a member from the leaderboard.
func (lb *Leaderboard) Remove(ctx context.Context, member string) error {
	return lb.client.ZRem(ctx, lb.key, member).Err()
}

// Score is a convenience to parse string scores; used by HTTP handlers.
func ParseScore(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
