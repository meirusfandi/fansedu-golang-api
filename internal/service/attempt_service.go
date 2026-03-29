package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AttemptService interface {
	Start(ctx context.Context, userID, tryoutSessionID string) (domain.Attempt, error)
	GetByID(ctx context.Context, attemptID, userID string) (domain.Attempt, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Attempt, error)
	Submit(ctx context.Context, attemptID, userID string) (domain.Attempt, *domain.AttemptFeedback, *TryoutSubmitAnalysis, error)
	// TryoutAnalysisForAttempt membangun ulang review + agregat modul (untuk GET attempt setelah submit).
	TryoutAnalysisForAttempt(ctx context.Context, attemptID, tryoutSessionID string) (*TryoutSubmitAnalysis, error)
}
