package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type DashboardSummary struct {
	TotalAttempts int     `json:"totalAttempts"`
	AvgScore      float64 `json:"avgScore"`
	AvgPercentile float64 `json:"avgPercentile"`
}

type DashboardResponse struct {
	Summary            DashboardSummary       `json:"summary"`
	OpenTryouts        []domain.TryoutSession `json:"openTryouts"`
	RecentAttempts     []domain.Attempt       `json:"recentAttempts"`
	StrengthAreas      []string               `json:"strengthAreas"`
	ImprovementAreas   []string               `json:"improvementAreas"`
	Recommendation     string                 `json:"recommendation"`
	LearningEvaluation *AttemptEvaluation     `json:"learningEvaluation,omitempty"`
}

type DashboardService interface {
	GetStudentDashboard(ctx context.Context, userID string) (*DashboardResponse, error)
}
