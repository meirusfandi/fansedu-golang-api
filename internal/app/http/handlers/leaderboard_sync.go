package handlers

import (
	"context"
	"log"

	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
)

// ReconcileTryoutLeaderboardRedis memuat skor terbaik per user dari DB (hanya yang terdaftar + submitted + score not null)
// dan mengganti ZSET Redis agar GET .../leaderboard/top dan /rank selaras dengan GET .../leaderboard (DB).
func ReconcileTryoutLeaderboardRedis(ctx context.Context, deps *Deps, tryoutID string) {
	if deps == nil || deps.Redis == nil || deps.TryoutRegistrationRepo == nil || tryoutID == "" {
		return
	}
	rows, err := deps.TryoutRegistrationRepo.ListLeaderboardRedisSyncRows(ctx, tryoutID)
	if err != nil {
		log.Printf("leaderboard redis reconcile: query tryout=%s err=%v", tryoutID, err)
		return
	}
	z := make([]cache.LeaderboardZMember, 0, len(rows))
	for _, row := range rows {
		z = append(z, cache.LeaderboardZMember{UserID: row.UserID, Score: row.Score})
	}
	if err := cache.ReplaceTryoutLeaderboardZSet(ctx, deps.Redis, tryoutID, z); err != nil {
		log.Printf("leaderboard redis reconcile: replace tryout=%s err=%v", tryoutID, err)
	}
}
