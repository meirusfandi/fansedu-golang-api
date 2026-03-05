package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type dashboardService struct {
	userRepo interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
	}
	attemptRepo interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	}
	tryoutRepo interface {
		ListOpenForStudent(ctx context.Context, now time.Time, subjectID *string) ([]domain.TryoutSession, error)
	}
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
	}
	questionRepo interface {
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
	}
	answerRepo interface {
		ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
	}
}

func NewDashboardService(
	userRepo interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
	},
	attemptRepo interface {
		ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	},
	tryoutRepo interface {
		ListOpenForStudent(ctx context.Context, now time.Time, subjectID *string) ([]domain.TryoutSession, error)
	},
	feedbackRepo interface {
		GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
	},
	questionRepo interface {
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
	},
	answerRepo interface {
		ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
	},
) DashboardService {
	return &dashboardService{
		userRepo:     userRepo,
		attemptRepo:  attemptRepo,
		tryoutRepo:   tryoutRepo,
		feedbackRepo: feedbackRepo,
		questionRepo: questionRepo,
		answerRepo:   answerRepo,
	}
}

func (s *dashboardService) GetStudentDashboard(ctx context.Context, userID string) (*DashboardResponse, error) {
	var subjectID *string
	if u, err := s.userRepo.FindByID(ctx, userID); err == nil {
		subjectID = u.SubjectID
	}
	openTryouts, _ := s.tryoutRepo.ListOpenForStudent(ctx, time.Now(), subjectID)
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
	resp := &DashboardResponse{
		Summary: DashboardSummary{
			TotalAttempts: n,
			AvgScore:      avgScore,
			AvgPercentile: avgPct,
		},
		OpenTryouts:      openTryouts,
		RecentAttempts:   attempts,
		StrengthAreas:    strengthAreas,
		ImprovementAreas: improvementAreas,
		Recommendation:   rec,
	}
	// Detail penilaian + rekomendasi dari attempt terakhir yang sudah submit
	var lastSubmitted *domain.Attempt
	for i := range attempts {
		if attempts[i].Status == domain.AttemptStatusSubmitted {
			lastSubmitted = &attempts[i]
			break
		}
	}
	if lastSubmitted != nil {
		questions, _ := s.questionRepo.ListByTryoutSessionID(ctx, lastSubmitted.TryoutSessionID)
		answers, _ := s.answerRepo.ListByAttemptID(ctx, lastSubmitted.ID)
		if len(questions) > 0 {
			eval := EvaluateAttemptAnswers(questions, answers)
			eval.AttemptID = lastSubmitted.ID
			// Gabung dengan feedback yang sudah ada jika ada
			if len(eval.StrengthAreas) > 0 {
				strengthAreas = eval.StrengthAreas
			}
			if len(eval.ImprovementAreas) > 0 {
				improvementAreas = eval.ImprovementAreas
			}
			if eval.Recommendation != "" {
				resp.Recommendation = eval.Recommendation
			}
			resp.StrengthAreas = strengthAreas
			resp.ImprovementAreas = improvementAreas
			resp.LearningEvaluation = &eval
		}
	}
	return resp, nil
}
