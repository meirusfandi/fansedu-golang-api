package service

import (
	"context"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type tryoutService struct {
	tryoutRepo   TryoutRepo
	registration TryoutRegistrationRepo
}

type TryoutRepo interface {
	ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
	ListOpenForStudent(ctx context.Context, now time.Time, subjectID *string) ([]domain.TryoutSession, error)
	ListForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error)
	GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
	Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	Update(ctx context.Context, t domain.TryoutSession) error
	Delete(ctx context.Context, id string) error
}

type TryoutRegistrationRepo interface {
	Register(ctx context.Context, userID, tryoutID string) error
	ListLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error)
	EnsureAllStudentsForTryout(ctx context.Context, tryoutID string) error
	EnsureStudentForAllOpenTryouts(ctx context.Context, userID string) error
}

func NewTryoutService(tryoutRepo TryoutRepo, registration TryoutRegistrationRepo) TryoutService {
	return &tryoutService{tryoutRepo: tryoutRepo, registration: registration}
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
	return s.registration.ListLeaderboard(ctx, tryoutID)
}
