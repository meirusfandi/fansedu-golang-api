package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

const leaderboardCacheKeyPrefix = "fansedu:leaderboard:tryout:v1:"

func leaderboardCacheKey(tryoutID string) string {
	return leaderboardCacheKeyPrefix + tryoutID
}

type tryoutService struct {
	tryoutRepo   TryoutRepo
	registration TryoutRegistrationRepo
	rdb          *redis.Client
	lbTTL        time.Duration
}

func NewTryoutService(tryoutRepo TryoutRepo, registration TryoutRegistrationRepo, rdb *redis.Client, leaderboardCacheTTL time.Duration) TryoutService {
	if leaderboardCacheTTL <= 0 {
		leaderboardCacheTTL = time.Hour
	}
	return &tryoutService{
		tryoutRepo:   tryoutRepo,
		registration: registration,
		rdb:          rdb,
		lbTTL:        leaderboardCacheTTL,
	}
}

func (s *tryoutService) ListOpen(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListOpen(ctx, time.Now())
}

func (s *tryoutService) ListOpenForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListOpenForStudent(ctx, time.Now(), subjectID)
}

func (s *tryoutService) ListForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListForStudent(ctx, subjectID)
}

func (s *tryoutService) GetByID(ctx context.Context, id string) (domain.TryoutSession, error) {
	return s.tryoutRepo.GetByID(ctx, id)
}

func (s *tryoutService) Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error) {
	return s.tryoutRepo.Create(ctx, t)
}

func (s *tryoutService) Update(ctx context.Context, t domain.TryoutSession) error {
	return s.tryoutRepo.Update(ctx, t)
}

func (s *tryoutService) Delete(ctx context.Context, id string) error {
	return s.tryoutRepo.Delete(ctx, id)
}

func (s *tryoutService) Register(ctx context.Context, userID, tryoutID string) error {
	return s.registration.Register(ctx, userID, tryoutID)
}

func (s *tryoutService) GetLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error) {
	if s.rdb != nil {
		key := leaderboardCacheKey(tryoutID)
		val, err := s.rdb.Get(ctx, key).Bytes()
		if err == nil && len(val) > 0 {
			var list []domain.LeaderboardEntry
			if json.Unmarshal(val, &list) == nil {
				return list, nil
			}
		}
	}

	list, err := s.registration.ListLeaderboard(ctx, tryoutID)
	if err != nil {
		return nil, err
	}
	if s.rdb != nil && list != nil {
		b, err := json.Marshal(list)
		if err == nil {
			_ = s.rdb.Set(ctx, leaderboardCacheKey(tryoutID), b, s.lbTTL).Err()
		}
	}
	return list, nil
}

func (s *tryoutService) InvalidateLeaderboardCache(ctx context.Context, tryoutID string) {
	if s.rdb == nil || tryoutID == "" {
		return
	}
	_ = s.rdb.Del(ctx, leaderboardCacheKey(tryoutID)).Err()
}
