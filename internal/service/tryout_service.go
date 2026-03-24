package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type TryoutService interface {
	ListOpen(ctx context.Context) ([]domain.TryoutSession, error)
	ListOpenForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error)
	ListForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error)
	GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
	Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	Update(ctx context.Context, t domain.TryoutSession) error
	Delete(ctx context.Context, id string) error
	Register(ctx context.Context, userID, tryoutID string) error
	GetLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error)
	// InvalidateLeaderboardCache hapus cache Redis untuk tryout (no-op jika Redis tidak dipakai).
	InvalidateLeaderboardCache(ctx context.Context, tryoutID string)
}
