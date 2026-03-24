package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// LeaderboardZAdd menyimpan / update skor user di sorted set (member=userID). Tanpa TTL (persistent di Redis).
func LeaderboardZAdd(ctx context.Context, rdb *redis.Client, tryoutID, userID string, score float64) error {
	if rdb == nil || tryoutID == "" || userID == "" {
		return nil
	}
	return rdb.ZAdd(ctx, LeaderboardZKey(tryoutID), redis.Z{
		Score:  score,
		Member: userID,
	}).Err()
}

// LeaderboardTop mengembalikan top N (rank tertinggi dulu).
func LeaderboardTop(ctx context.Context, rdb *redis.Client, tryoutID string, n int) ([]redis.Z, error) {
	if rdb == nil || tryoutID == "" || n <= 0 {
		return nil, nil
	}
	if n > 1000 {
		n = 1000
	}
	return rdb.ZRevRangeWithScores(ctx, LeaderboardZKey(tryoutID), 0, int64(n-1)).Result()
}

// LeaderboardUserRankScore: rank 0-based dari atas (0 = juara 1), skor, ok=false jika user belum punya entri di ZSET.
func LeaderboardUserRankScore(ctx context.Context, rdb *redis.Client, tryoutID, userID string) (rank int64, score float64, ok bool, err error) {
	if rdb == nil || tryoutID == "" || userID == "" {
		return 0, 0, false, nil
	}
	key := LeaderboardZKey(tryoutID)
	score, err = rdb.ZScore(ctx, key, userID).Result()
	if err == redis.Nil {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	rank, err = rdb.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return 0, 0, false, err
	}
	return rank, score, true, nil
}
