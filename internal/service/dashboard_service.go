package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type DashboardSummary struct {
	TotalAttempts  int     `json:"total_attempts"`
	AvgScore       float64 `json:"avg_score"`
	AvgPercentile  float64 `json:"avg_percentile"`
}

type DashboardResponse struct {
	Summary            DashboardSummary       `json:"summary"`
	OpenTryouts        []domain.TryoutSession `json:"open_tryouts"`
	RecentAttempts     []domain.Attempt       `json:"recent_attempts"`
	StrengthAreas      []string               `json:"strength_areas"`
	ImprovementAreas   []string               `json:"improvement_areas"`
	Recommendation     string                 `json:"recommendation"`
	LearningEvaluation *AttemptEvaluation     `json:"learning_evaluation,omitempty"`
}

type DashboardService interface {
	GetStudentDashboard(ctx context.Context, userID string) (*DashboardResponse, error)
}
