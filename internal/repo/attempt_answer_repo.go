package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AttemptAnswerRepo interface {
	Upsert(ctx context.Context, a domain.AttemptAnswer) error
	GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error)
	ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
}

type attemptAnswerRepo struct{ pool *pgxpool.Pool }

func NewAttemptAnswerRepo(pool *pgxpool.Pool) AttemptAnswerRepo { return &attemptAnswerRepo{pool: pool} }

func (r *attemptAnswerRepo) Upsert(ctx context.Context, a domain.AttemptAnswer) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attempt_answers (id, attempt_id, question_id, answer_text, selected_option, is_marked)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6)
		ON CONFLICT (attempt_id, question_id) DO UPDATE SET answer_text = EXCLUDED.answer_text, selected_option = EXCLUDED.selected_option, is_marked = EXCLUDED.is_marked, updated_at = NOW()
	`, id, a.AttemptID, a.QuestionID, a.AnswerText, a.SelectedOption, a.IsMarked)
	return err
}

func (r *attemptAnswerRepo) GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, attempt_id, question_id, answer_text, selected_option, is_marked, created_at, updated_at
		FROM attempt_answers WHERE attempt_id = $1::uuid AND question_id = $2::uuid
	`, attemptID, questionID)
	var a domain.AttemptAnswer
	err := row.Scan(&a.ID, &a.AttemptID, &a.QuestionID, &a.AnswerText, &a.SelectedOption, &a.IsMarked, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r *attemptAnswerRepo) ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, attempt_id, question_id, answer_text, selected_option, is_marked, created_at, updated_at
		FROM attempt_answers WHERE attempt_id = $1::uuid
	`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.AttemptAnswer
	for rows.Next() {
		var a domain.AttemptAnswer
		if err := rows.Scan(&a.ID, &a.AttemptID, &a.QuestionID, &a.AnswerText, &a.SelectedOption, &a.IsMarked, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
