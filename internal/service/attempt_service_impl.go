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
	Update(ctx context.Context, a domain.Attempt) error
}

type AttemptAnswerRepo interface {
	Upsert(ctx context.Context, a domain.AttemptAnswer) error
	ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
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

func (s *attemptService) Submit(ctx context.Context, attemptID, userID string) (domain.Attempt, *domain.AttemptFeedback, error) {
	a, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return domain.Attempt{}, nil, ErrAttemptNotFound
	}
	if a.UserID != userID {
		return domain.Attempt{}, nil, ErrNotYourAttempt
	}
	if a.Status == domain.AttemptStatusSubmitted {
		return a, nil, ErrAlreadySubmitted
	}
	answers, _ := s.answerRepo.ListByAttemptID(ctx, attemptID)
	questions, _ := s.questionRepo.ListByTryoutSessionID(ctx, a.TryoutSessionID)
	score, maxScore := computeScore(questions, answers)
	percentile := float64(0)
	// TODO: compute percentile from all attempts for this tryout
	now := time.Now()
	a.SubmittedAt = &now
	a.Status = domain.AttemptStatusSubmitted
	a.Score = &score
	a.MaxScore = &maxScore
	a.Percentile = &percentile
	// Waktu pengerjaan (detik) untuk leaderboard
	if sec := int(now.Sub(a.StartedAt).Seconds()); sec >= 0 {
		a.TimeSecondsSpent = &sec
	}
	if err := s.attemptRepo.Update(ctx, a); err != nil {
		return domain.Attempt{}, nil, err
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
		return a, nil, nil
	}
	return a, &fb, nil
}

func computeScore(questions []domain.Question, answers []domain.AttemptAnswer) (float64, float64) {
	answerMap := make(map[string]domain.AttemptAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}
	var score, maxScore float64
	for _, q := range questions {
		maxScore += q.MaxScore
		ans, ok := answerMap[q.ID]
		if !ok {
			continue
		}
		var got float64
		switch q.Type {
		case domain.QuestionTypeShort:
			if ans.AnswerText != nil && *ans.AnswerText != "" {
				got = q.MaxScore * 0.5
			}
		case domain.QuestionTypeMultipleChoice, domain.QuestionTypeTrueFalse:
			if ans.SelectedOption != nil && *ans.SelectedOption != "" {
				got = q.MaxScore
			}
		}
		score += got
	}
	return score, maxScore
}

func formatScore(f float64) string { return fmt.Sprintf("%.2f", f) }
func strPtr(s string) *string     { return &s }
