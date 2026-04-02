package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AttemptAnswerRepo interface {
	Upsert(ctx context.Context, a domain.AttemptAnswer) error
	GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error)
	ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error)
	ListByQuestionFromSubmittedAttempts(ctx context.Context, tryoutSessionID, questionID string) ([]domain.AttemptAnswer, error)
	ListByTryoutFromSubmittedAttempts(ctx context.Context, tryoutSessionID string) ([]domain.AttemptAnswer, error)
	SetAnswerGrading(ctx context.Context, attemptID, questionID string, isCorrect *bool) error
	UpdateAnswerReview(ctx context.Context, attemptID, questionID string, reviewerComment *string, manualScore *float64, reviewedByUserID string) error
	UpdateManualScoreReview(ctx context.Context, attemptID, questionID string, manualScore *float64, reviewedByUserID string) error
	EnsureAnswerRowForReview(ctx context.Context, attemptID, questionID string) error
	// ClearManualGradingForAttempt hapus manual_score; jika clearReviewMeta, hapus juga komentar & metadata review.
	ClearManualGradingForAttempt(ctx context.Context, attemptID string, clearReviewMeta bool) error
}

type attemptAnswerRepo struct{ pool *pgxpool.Pool }

func NewAttemptAnswerRepo(pool *pgxpool.Pool) AttemptAnswerRepo { return &attemptAnswerRepo{pool: pool} }

func scanAttemptAnswer(row interface {
	Scan(dest ...any) error
}) (domain.AttemptAnswer, error) {
	var a domain.AttemptAnswer
	var ms sql.NullFloat64
	var rc, rb sql.NullString
	var rat sql.NullTime
	err := row.Scan(
		&a.ID, &a.AttemptID, &a.QuestionID,
		&a.AnswerText, &a.SelectedOption, &a.IsMarked, &a.IsCorrect,
		&ms, &rc, &rb, &rat,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return a, err
	}
	if ms.Valid {
		v := ms.Float64
		a.ManualScore = &v
	}
	if rc.Valid {
		v := rc.String
		a.ReviewerComment = &v
	}
	if rb.Valid {
		v := rb.String
		a.ReviewedByUserID = &v
	}
	if rat.Valid {
		t := rat.Time
		a.ReviewedAt = &t
	}
	return a, nil
}

func (r *attemptAnswerRepo) Upsert(ctx context.Context, a domain.AttemptAnswer) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attempt_answers (id, attempt_id, question_id, answer_text, selected_option, is_marked)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6)
		ON CONFLICT (attempt_id, question_id) DO UPDATE SET
			answer_text = EXCLUDED.answer_text,
			selected_option = EXCLUDED.selected_option,
			is_marked = EXCLUDED.is_marked,
			updated_at = NOW()
	`, id, a.AttemptID, a.QuestionID, a.AnswerText, a.SelectedOption, a.IsMarked)
	return err
}

func (r *attemptAnswerRepo) SetAnswerGrading(ctx context.Context, attemptID, questionID string, isCorrect *bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attempt_answers SET is_correct = $3, updated_at = NOW()
		WHERE attempt_id = $1::uuid AND question_id = $2::uuid
	`, attemptID, questionID, isCorrect)
	return err
}

func (r *attemptAnswerRepo) UpdateAnswerReview(ctx context.Context, attemptID, questionID string, reviewerComment *string, manualScore *float64, reviewedByUserID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attempt_answers SET
			reviewer_comment = $1,
			manual_score = $2,
			reviewed_by_user_id = $3::uuid,
			reviewed_at = NOW(),
			updated_at = NOW()
		WHERE attempt_id = $4::uuid AND question_id = $5::uuid
	`, reviewerComment, manualScore, reviewedByUserID, attemptID, questionID)
	return err
}

// UpdateManualScoreReview khusus update manual_score agar tidak menimpa field review lain saat save cepat berulang.
func (r *attemptAnswerRepo) UpdateManualScoreReview(ctx context.Context, attemptID, questionID string, manualScore *float64, reviewedByUserID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attempt_answers SET
			manual_score = $1,
			reviewed_by_user_id = $2::uuid,
			reviewed_at = NOW(),
			updated_at = NOW()
		WHERE attempt_id = $3::uuid AND question_id = $4::uuid
	`, manualScore, reviewedByUserID, attemptID, questionID)
	return err
}

func (r *attemptAnswerRepo) EnsureAnswerRowForReview(ctx context.Context, attemptID, questionID string) error {
	id := uuid.New().String()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO attempt_answers (id, attempt_id, question_id, is_marked)
		VALUES ($1::uuid, $2::uuid, $3::uuid, false)
		ON CONFLICT (attempt_id, question_id) DO NOTHING
	`, id, attemptID, questionID)
	return err
}

func (r *attemptAnswerRepo) ClearManualGradingForAttempt(ctx context.Context, attemptID string, clearReviewMeta bool) error {
	if clearReviewMeta {
		_, err := r.pool.Exec(ctx, `
			UPDATE attempt_answers SET
				manual_score = NULL,
				reviewer_comment = NULL,
				reviewed_by_user_id = NULL,
				reviewed_at = NULL,
				updated_at = NOW()
			WHERE attempt_id = $1::uuid
		`, attemptID)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE attempt_answers SET manual_score = NULL, updated_at = NOW()
		WHERE attempt_id = $1::uuid
	`, attemptID)
	return err
}

func (r *attemptAnswerRepo) GetByAttemptAndQuestion(ctx context.Context, attemptID, questionID string) (domain.AttemptAnswer, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, attempt_id, question_id, answer_text, selected_option, is_marked, is_correct,
			manual_score, reviewer_comment, reviewed_by_user_id, reviewed_at, created_at, updated_at
		FROM attempt_answers WHERE attempt_id = $1::uuid AND question_id = $2::uuid
	`, attemptID, questionID)
	return scanAttemptAnswer(row)
}

func (r *attemptAnswerRepo) ListByAttemptID(ctx context.Context, attemptID string) ([]domain.AttemptAnswer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, attempt_id, question_id, answer_text, selected_option, is_marked, is_correct,
			manual_score, reviewer_comment, reviewed_by_user_id, reviewed_at, created_at, updated_at
		FROM attempt_answers WHERE attempt_id = $1::uuid
	`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.AttemptAnswer
	for rows.Next() {
		a, err := scanAttemptAnswer(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *attemptAnswerRepo) ListByQuestionFromSubmittedAttempts(ctx context.Context, tryoutSessionID, questionID string) ([]domain.AttemptAnswer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT aa.id, aa.attempt_id, aa.question_id, aa.answer_text, aa.selected_option, aa.is_marked, aa.is_correct,
			aa.manual_score, aa.reviewer_comment, aa.reviewed_by_user_id, aa.reviewed_at, aa.created_at, aa.updated_at
		FROM attempt_answers aa
		INNER JOIN attempts a ON a.id = aa.attempt_id
		WHERE a.tryout_session_id = $1::uuid AND a.status = 'submitted' AND aa.question_id = $2::uuid
	`, tryoutSessionID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.AttemptAnswer
	for rows.Next() {
		a, err := scanAttemptAnswer(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *attemptAnswerRepo) ListByTryoutFromSubmittedAttempts(ctx context.Context, tryoutSessionID string) ([]domain.AttemptAnswer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT aa.id, aa.attempt_id, aa.question_id, aa.answer_text, aa.selected_option, aa.is_marked, aa.is_correct,
			aa.manual_score, aa.reviewer_comment, aa.reviewed_by_user_id, aa.reviewed_at, aa.created_at, aa.updated_at
		FROM attempt_answers aa
		INNER JOIN attempts a ON a.id = aa.attempt_id
		WHERE a.tryout_session_id = $1::uuid AND a.status = 'submitted'
	`, tryoutSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.AttemptAnswer
	for rows.Next() {
		a, err := scanAttemptAnswer(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}
