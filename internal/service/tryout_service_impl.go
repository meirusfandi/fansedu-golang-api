package service

import (
	"context"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

type tryoutService struct {
	tryoutRepo   repo.TryoutRepo
	registration repo.TryoutRegistrationRepo
}

func NewTryoutService(tryoutRepo repo.TryoutRepo, registration repo.TryoutRegistrationRepo) TryoutService {
	return &tryoutService{tryoutRepo: tryoutRepo, registration: registration}
}

func (s *tryoutService) List(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.List(ctx)
}

func (s *tryoutService) ListOpen(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListOpen(ctx, time.Now())
}

func (s *tryoutService) ListOpenForStudent(ctx context.Context, subjectID *string, levelID *string) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListOpenForStudent(ctx, time.Now(), subjectID, levelID)
}

func (s *tryoutService) ListForStudent(ctx context.Context, subjectID *string, levelID *string) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListForStudent(ctx, subjectID, levelID)
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
	return s.registration.ListLeaderboard(ctx, tryoutID)
}
