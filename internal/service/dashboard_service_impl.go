package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type dashboardService struct {
	attemptRepo  interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	}
	tryoutRepo interface {
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
	}
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
	}
}

func NewDashboardService(
	attemptRepo interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	},
	tryoutRepo interface {
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
	},
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
	},
) DashboardService {
	return &dashboardService{
		attemptRepo:  attemptRepo,
		tryoutRepo:   tryoutRepo,
		feedbackRepo: feedbackRepo,
	}
}

func (s *dashboardService) GetStudentDashboard(ctx context.Context, userID string) (*DashboardResponse, error) {
	openTryouts, _ := s.tryoutRepo.ListOpen(ctx, time.Now())
	attempts, _ := s.attemptRepo.ListByUserID(ctx, userID)
	if attempts == nil {
		attempts = []domain.Attempt{}
	}
	const recentN = 5
	if len(attempts) > recentN {
		attempts = attempts[:recentN]
	}
	var totalScore, totalPct float64
	var strengthAreas, improvementAreas []string
	for _, a := range attempts {
		if a.Score != nil {
			totalScore += *a.Score
		}
		if a.Percentile != nil {
			totalPct += *a.Percentile
		}
		fb, _ := s.feedbackRepo.GetByAttemptID(ctx, a.ID)
		if len(fb.StrengthAreas) > 0 {
			_ = json.Unmarshal(fb.StrengthAreas, &strengthAreas)
		}
		if len(fb.ImprovementAreas) > 0 {
			_ = json.Unmarshal(fb.ImprovementAreas, &improvementAreas)
		}
	}
	n := len(attempts)
	avgScore, avgPct := 0.0, 0.0
	if n > 0 {
		avgScore = totalScore / float64(n)
		avgPct = totalPct / float64(n)
	}
	rec := "Mulai tryout untuk melihat rekomendasi."
	if n > 0 {
		rec = "Lanjutkan berlatih dan perbaiki area yang masih lemah."
	}
	return &DashboardResponse{
		Summary: DashboardSummary{
			TotalAttempts:  n,
			AvgScore:       avgScore,
			AvgPercentile:  avgPct,
		},
		OpenTryouts:      openTryouts,
		RecentAttempts:   attempts,
		StrengthAreas:    strengthAreas,
		ImprovementAreas: improvementAreas,
		Recommendation:   rec,
	}, nil
}
