package service

import (
	"context"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type tryoutService struct {
	tryoutRepo TryoutRepo
}

type TryoutRepo interface {
	ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
	GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
	Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	Update(ctx context.Context, t domain.TryoutSession) error
	Delete(ctx context.Context, id string) error
}

func NewTryoutService(tryoutRepo TryoutRepo) TryoutService {
	return &tryoutService{tryoutRepo: tryoutRepo}
}

func (s *tryoutService) ListOpen(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.ListOpen(ctx, time.Now())
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
