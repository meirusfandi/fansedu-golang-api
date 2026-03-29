package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/ai"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

var (
	ErrAttemptNotFound   = errors.New("attempt not found")
	ErrNotYourAttempt    = errors.New("attempt does not belong to you")
	ErrAlreadySubmitted  = errors.New("attempt already submitted")
)

type attemptService struct {
	attemptRepo       AttemptRepo
	answerRepo        AttemptAnswerRepo
	feedbackRepo      FeedbackRepo
	questionRepo      QuestionRepo
	tryoutRepo        repo.TryoutRepo
	feedbackGenerator ai.FeedbackGenerator
}

type AttemptRepo interface {
	Create(ctx context.Context, a domain.Attempt) (domain.Attempt, error)
	GetByID(ctx context.Context, id string) (domain.Attempt, error)
	GetByUserAndTryout(ctx context.Context, userID, tryoutSessionID string) (domain.Attempt, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	ListSubmittedByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Attempt, error)
	Update(ctx context.Context, a domain.Attempt) error
}

type AttemptAnswerRepo interface {
	Upsert(ctx context.Context, a domain.AttemptAnswer) error
	ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
	SetAnswerGrading(ctx context.Context, attemptID, questionID string, isCorrect *bool) error
}

type FeedbackRepo interface {
	Create(ctx context.Context, f domain.AttemptFeedback) (domain.AttemptFeedback, error)
	GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
}

type QuestionRepo interface {
	ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
	GetByID(ctx context.Context, id string) (domain.Question, error)
}

func NewAttemptService(attemptRepo AttemptRepo, answerRepo AttemptAnswerRepo, feedbackRepo FeedbackRepo, questionRepo QuestionRepo, tryoutRepo repo.TryoutRepo, feedbackGenerator ai.FeedbackGenerator) AttemptService {
	if feedbackGenerator == nil {
		feedbackGenerator = ai.NewFallbackFeedbackGenerator()
	}
	return &attemptService{
		attemptRepo:       attemptRepo,
		answerRepo:        answerRepo,
		feedbackRepo:      feedbackRepo,
		questionRepo:      questionRepo,
		tryoutRepo:        tryoutRepo,
		feedbackGenerator: feedbackGenerator,
	}
}

func (s *attemptService) Start(ctx context.Context, userID, tryoutSessionID string) (domain.Attempt, error) {
	_, err := s.tryoutRepo.GetByID(ctx, tryoutSessionID)
	if err != nil {
		return domain.Attempt{}, err
	}
	existing, err := s.attemptRepo.GetByUserAndTryout(ctx, userID, tryoutSessionID)
	if err == nil {
		if existing.Status == domain.AttemptStatusSubmitted {
			return domain.Attempt{}, ErrAlreadySubmitted
		}
		return existing, nil
	}
	a := domain.Attempt{
		UserID:          userID,
		TryoutSessionID:  tryoutSessionID,
		Status:          domain.AttemptStatusInProgress,
	}
	return s.attemptRepo.Create(ctx, a)
}

func (s *attemptService) GetByID(ctx context.Context, attemptID, userID string) (domain.Attempt, error) {
	a, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return domain.Attempt{}, ErrAttemptNotFound
	}
	if a.UserID != userID {
		return domain.Attempt{}, ErrNotYourAttempt
	}
	return a, nil
}

func (s *attemptService) ListByUser(ctx context.Context, userID string) ([]domain.Attempt, error) {
	return s.attemptRepo.ListByUserID(ctx, userID)
}

func (s *attemptService) Submit(ctx context.Context, attemptID, userID string) (domain.Attempt, *domain.AttemptFeedback, *TryoutSubmitAnalysis, error) {
	a, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return domain.Attempt{}, nil, nil, ErrAttemptNotFound
	}
	if a.UserID != userID {
		return domain.Attempt{}, nil, nil, ErrNotYourAttempt
	}
	if a.Status == domain.AttemptStatusSubmitted {
		return a, nil, nil, ErrAlreadySubmitted
	}
	answers, _ := s.answerRepo.ListByAttemptID(ctx, attemptID)
	questions, _ := s.questionRepo.ListByTryoutSessionID(ctx, a.TryoutSessionID)
	score, maxScore, outcomes, modAggs := GradeTryoutAttempt(questions, answers)
	for _, o := range outcomes {
		_ = s.answerRepo.SetAnswerGrading(ctx, attemptID, o.QuestionID, o.IsCorrect)
	}
	analysis := &TryoutSubmitAnalysis{Review: outcomes, Modules: modAggs}
	// Persentil: hanya jika ada ≥2 skor (peserta lain + attempt ini); jika tidak, NULL (bukan 0 palsu).
	var percentile *float64
	if others, perr := s.attemptRepo.ListSubmittedByTryoutSessionID(ctx, a.TryoutSessionID); perr == nil {
		scores := make([]float64, 0, len(others)+1)
		for _, o := range others {
			if o.Score != nil {
				scores = append(scores, *o.Score)
			}
		}
		scores = append(scores, score)
		percentile = percentileRankPercent(scores, score)
	}
	now := time.Now()
	a.SubmittedAt = &now
	a.Status = domain.AttemptStatusSubmitted
	a.Score = &score
	a.MaxScore = &maxScore
	a.Percentile = percentile
	// Waktu pengerjaan (detik) untuk leaderboard
	if sec := int(now.Sub(a.StartedAt).Seconds()); sec >= 0 {
		a.TimeSecondsSpent = &sec
	}
	if err := s.attemptRepo.Update(ctx, a); err != nil {
		return domain.Attempt{}, nil, nil, err
	}
	// Generate feedback berdasarkan jawaban (AI atau fallback), lalu simpan ke attempt_feedback
	gen, err := s.feedbackGenerator.Generate(ctx, ai.FeedbackRequest{
		Questions: questions,
		Answers:   answers,
		Score:     score,
		MaxScore:  maxScore,
	})
	if err != nil {
		gen = &ai.GeneratedFeedback{
			Summary:          "Tryout selesai.",
			Recap:            "Skor Anda: " + formatScore(score) + " dari " + formatScore(maxScore) + ".",
			StrengthAreas:    []string{},
			ImprovementAreas: []string{"Lanjutkan berlatih."},
			Recommendation:   "Lanjutkan berlatih dan perbaiki area yang masih lemah.",
		}
	}
	strength, _ := json.Marshal(gen.StrengthAreas)
	improvement, _ := json.Marshal(gen.ImprovementAreas)
	fb := domain.AttemptFeedback{
		AttemptID:          attemptID,
		Summary:            &gen.Summary,
		Recap:              &gen.Recap,
		StrengthAreas:      strength,
		ImprovementAreas:   improvement,
		RecommendationText: &gen.Recommendation,
	}
	fb, err = s.feedbackRepo.Create(ctx, fb)
	if err != nil {
		return a, nil, analysis, nil
	}
	return a, &fb, analysis, nil
}

func (s *attemptService) TryoutAnalysisForAttempt(ctx context.Context, attemptID, tryoutSessionID string) (*TryoutSubmitAnalysis, error) {
	answers, err := s.answerRepo.ListByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	questions, err := s.questionRepo.ListByTryoutSessionID(ctx, tryoutSessionID)
	if err != nil {
		return nil, err
	}
	_, _, outcomes, modAggs := GradeTryoutAttempt(questions, answers)
	return &TryoutSubmitAnalysis{Review: outcomes, Modules: modAggs}, nil
}

func formatScore(f float64) string { return fmt.Sprintf("%.2f", f) }
func strPtr(s string) *string     { return &s }
