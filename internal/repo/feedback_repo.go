package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type FeedbackRepo interface {
	Create(ctx context.Context, f domain.AttemptFeedback) (domain.AttemptFeedback, error)
	GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error)
	Update(ctx context.Context, f domain.AttemptFeedback) error
}

type feedbackRepo struct{ pool *pgxpool.Pool }

func NewFeedbackRepo(pool *pgxpool.Pool) FeedbackRepo { return &feedbackRepo{pool: pool} }

func (r *feedbackRepo) Create(ctx context.Context, f domain.AttemptFeedback) (domain.AttemptFeedback, error) {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attempt_feedback (id, attempt_id, summary, recap, strength_areas, improvement_areas, recommendation_text)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
	`, id, f.AttemptID, f.Summary, f.Recap, f.StrengthAreas, f.ImprovementAreas, f.RecommendationText)
	if err != nil {
		return domain.AttemptFeedback{}, err
	}
	f.ID = id
	return f, nil
}

func (r *feedbackRepo) GetByAttemptID(ctx context.Context, attemptID string) (domain.AttemptFeedback, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, attempt_id, summary, recap, strength_areas, improvement_areas, recommendation_text, created_at, updated_at
		FROM attempt_feedback WHERE attempt_id = $1::uuid
	`, attemptID)
	var f domain.AttemptFeedback
	err := row.Scan(&f.ID, &f.AttemptID, &f.Summary, &f.Recap, &f.StrengthAreas, &f.ImprovementAreas, &f.RecommendationText, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func (r *feedbackRepo) Update(ctx context.Context, f domain.AttemptFeedback) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attempt_feedback SET summary=$2, recap=$3, strength_areas=$4, improvement_areas=$5, recommendation_text=$6
		WHERE id = $1::uuid
	`, f.ID, f.Summary, f.Recap, f.StrengthAreas, f.ImprovementAreas, f.RecommendationText)
	return err
}
