package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AttemptRepo interface {
	Create(ctx context.Context, a domain.Attempt) (domain.Attempt, error)
	GetByID(ctx context.Context, id string) (domain.Attempt, error)
	GetByUserAndTryout(ctx context.Context, userID, tryoutSessionID string) (domain.Attempt, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error)
	ListSubmittedByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Attempt, error)
	Update(ctx context.Context, a domain.Attempt) error
	AvgScoreSubmitted(ctx context.Context) (float64, error)
	ParticipantsCountByTryout(ctx context.Context, tryoutSessionID string) (int, error)
}

type attemptRepo struct{ pool *pgxpool.Pool }

func NewAttemptRepo(pool *pgxpool.Pool) AttemptRepo { return &attemptRepo{pool: pool} }

func (r *attemptRepo) Create(ctx context.Context, a domain.Attempt) (domain.Attempt, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO attempts (id, user_id, tryout_session_id, status)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::attempt_status)
		RETURNING id, user_id, tryout_session_id, started_at, submitted_at, status, score, max_score, percentile, time_seconds_spent, created_at, updated_at
	`, id, a.UserID, a.TryoutSessionID, a.Status).Scan(
		&a.ID, &a.UserID, &a.TryoutSessionID, &a.StartedAt, &a.SubmittedAt, &a.Status, &a.Score, &a.MaxScore, &a.Percentile, &a.TimeSecondsSpent, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return domain.Attempt{}, err
	}
	return a, nil
}

func (r *attemptRepo) GetByID(ctx context.Context, id string) (domain.Attempt, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, tryout_session_id, started_at, submitted_at, status, score, max_score, percentile, time_seconds_spent, created_at, updated_at
		FROM attempts WHERE id = $1::uuid
	`, id)
	var a domain.Attempt
	err := row.Scan(&a.ID, &a.UserID, &a.TryoutSessionID, &a.StartedAt, &a.SubmittedAt, &a.Status, &a.Score, &a.MaxScore, &a.Percentile, &a.TimeSecondsSpent, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r *attemptRepo) GetByUserAndTryout(ctx context.Context, userID, tryoutSessionID string) (domain.Attempt, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, tryout_session_id, started_at, submitted_at, status, score, max_score, percentile, time_seconds_spent, created_at, updated_at
		FROM attempts WHERE user_id = $1::uuid AND tryout_session_id = $2::uuid
	`, userID, tryoutSessionID)
	var a domain.Attempt
	err := row.Scan(&a.ID, &a.UserID, &a.TryoutSessionID, &a.StartedAt, &a.SubmittedAt, &a.Status, &a.Score, &a.MaxScore, &a.Percentile, &a.TimeSecondsSpent, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (r *attemptRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Attempt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, tryout_session_id, started_at, submitted_at, status, score, max_score, percentile, time_seconds_spent, created_at, updated_at
		FROM attempts WHERE user_id = $1::uuid ORDER BY started_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Attempt
	for rows.Next() {
		var a domain.Attempt
		if err := rows.Scan(&a.ID, &a.UserID, &a.TryoutSessionID, &a.StartedAt, &a.SubmittedAt, &a.Status, &a.Score, &a.MaxScore, &a.Percentile, &a.TimeSecondsSpent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *attemptRepo) ListSubmittedByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Attempt, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, tryout_session_id, started_at, submitted_at, status, score, max_score, percentile, time_seconds_spent, created_at, updated_at
		FROM attempts WHERE tryout_session_id = $1::uuid AND status = 'submitted' ORDER BY submitted_at DESC
	`, tryoutSessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Attempt
	for rows.Next() {
		var a domain.Attempt
		if err := rows.Scan(&a.ID, &a.UserID, &a.TryoutSessionID, &a.StartedAt, &a.SubmittedAt, &a.Status, &a.Score, &a.MaxScore, &a.Percentile, &a.TimeSecondsSpent, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *attemptRepo) AvgScoreSubmitted(ctx context.Context) (float64, error) {
	var avg *float64
	err := r.pool.QueryRow(ctx, `SELECT AVG(score) FROM attempts WHERE status = 'submitted'`).Scan(&avg)
	if err != nil || avg == nil {
		return 0, err
	}
	return *avg, nil
}

func (r *attemptRepo) Update(ctx context.Context, a domain.Attempt) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE attempts SET submitted_at=$2, status=$3::attempt_status, score=$4, max_score=$5, percentile=$6, time_seconds_spent=$7
		WHERE id = $1::uuid
	`, a.ID, a.SubmittedAt, a.Status, a.Score, a.MaxScore, a.Percentile, a.TimeSecondsSpent)
	return err
}

func (r *attemptRepo) ParticipantsCountByTryout(ctx context.Context, tryoutSessionID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM attempts
		WHERE tryout_session_id = $1::uuid AND status = 'submitted'
	`, tryoutSessionID).Scan(&n)
	return n, err
}
