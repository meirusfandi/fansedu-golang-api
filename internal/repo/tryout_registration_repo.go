package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type TryoutRegistrationRepo interface {
	Register(ctx context.Context, userID, tryoutID string) error
	IsRegistered(ctx context.Context, userID, tryoutID string) (bool, error)
	ListLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error)
	EnsureAllStudentsForTryout(ctx context.Context, tryoutID string) error
	EnsureStudentForAllOpenTryouts(ctx context.Context, userID string) error
}

type tryoutRegistrationRepo struct{ pool *pgxpool.Pool }

func NewTryoutRegistrationRepo(pool *pgxpool.Pool) TryoutRegistrationRepo {
	return &tryoutRegistrationRepo{pool: pool}
}

func (r *tryoutRegistrationRepo) Register(ctx context.Context, userID, tryoutID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tryout_registrations (user_id, tryout_session_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (user_id, tryout_session_id) DO NOTHING
	`, userID, tryoutID)
	return err
}

func (r *tryoutRegistrationRepo) IsRegistered(ctx context.Context, userID, tryoutID string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT 1 FROM tryout_registrations WHERE user_id = $1::uuid AND tryout_session_id = $2::uuid
	`, userID, tryoutID).Scan(&n)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListLeaderboard: nama siswa, sekolah, nilai. Urutan: nilai tertinggi DESC, waktu tercepat ASC, nama ASC; belum mengerjakan = nama ASC di akhir.
func (r *tryoutRegistrationRepo) ListLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error) {
	query := `
		WITH best_attempt AS (
			SELECT DISTINCT ON (a.user_id)
				a.user_id,
				a.score AS best_score,
				a.time_seconds_spent AS best_time_seconds
			FROM attempts a
			WHERE a.tryout_session_id = $1::uuid AND a.status = 'submitted' AND a.score IS NOT NULL
			ORDER BY a.user_id, a.score DESC, a.time_seconds_spent ASC NULLS LAST
		)
		SELECT
			u.id AS user_id,
			u.name AS user_name,
			s.name AS school_name,
			b.best_score,
			b.best_time_seconds
		FROM tryout_registrations r
		JOIN users u ON u.id = r.user_id
		LEFT JOIN schools s ON s.id = u.school_id
		LEFT JOIN best_attempt b ON b.user_id = r.user_id
		WHERE r.tryout_session_id = $1::uuid
		ORDER BY b.best_score DESC NULLS LAST, b.best_time_seconds ASC NULLS LAST, u.name ASC
	`
	rows, err := r.pool.Query(ctx, query, tryoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.LeaderboardEntry
	var rank int
	for rows.Next() {
		var e domain.LeaderboardEntry
		var bestScore *float64
		var bestTime *int
		var schoolName *string
		if err := rows.Scan(&e.UserID, &e.UserName, &schoolName, &bestScore, &bestTime); err != nil {
			return nil, err
		}
		rank++
		e.Rank = rank
		e.SchoolName = schoolName
		e.BestScore = bestScore
		e.BestTimeSeconds = bestTime
		e.HasAttempt = bestScore != nil
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *tryoutRegistrationRepo) EnsureAllStudentsForTryout(ctx context.Context, tryoutID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tryout_registrations (user_id, tryout_session_id)
		SELECT id, $1::uuid FROM users WHERE role = 'student'
		ON CONFLICT (user_id, tryout_session_id) DO NOTHING
	`, tryoutID)
	return err
}

func (r *tryoutRegistrationRepo) EnsureStudentForAllOpenTryouts(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tryout_registrations (user_id, tryout_session_id)
		SELECT $1::uuid, id FROM tryout_sessions WHERE status != 'draft'
		ON CONFLICT (user_id, tryout_session_id) DO NOTHING
	`, userID)
	return err
}
