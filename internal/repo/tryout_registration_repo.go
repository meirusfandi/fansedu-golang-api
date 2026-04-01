package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// LeaderboardRedisSyncRow — satu baris untuk menyelaraskan Redis leaderboard dengan DB (skor attempt terbaik per user).
type LeaderboardRedisSyncRow struct {
	UserID string
	Score  float64
}

type TryoutRegistrationRepo interface {
	Register(ctx context.Context, userID, tryoutID string) error
	IsRegistered(ctx context.Context, userID, tryoutID string) (bool, error)
	GetRegisteredAt(ctx context.Context, userID, tryoutID string) (time.Time, bool, error)
	CountRegisteredForStudent(ctx context.Context, userID string, subjectID *string) (int, error)
	ListLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error)
	ListLeaderboardRedisSyncRows(ctx context.Context, tryoutID string) ([]LeaderboardRedisSyncRow, error)
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

func (r *tryoutRegistrationRepo) GetRegisteredAt(ctx context.Context, userID, tryoutID string) (time.Time, bool, error) {
	var t time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT registered_at
		FROM tryout_registrations
		WHERE user_id = $1::uuid AND tryout_session_id = $2::uuid
	`, userID, tryoutID).Scan(&t)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	return t, true, nil
}

func (r *tryoutRegistrationRepo) CountRegisteredForStudent(ctx context.Context, userID string, subjectID *string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tryout_registrations tr
		JOIN tryout_sessions t ON t.id = tr.tryout_session_id
		WHERE tr.user_id = $1::uuid
		  AND (
			$2::uuid IS NULL
			OR t.subject_id IS NULL
			OR t.subject_id = $2::uuid
		  )
	`
	var subj interface{} = nil
	if subjectID != nil && *subjectID != "" {
		subj = *subjectID
	}
	var n int
	err := r.pool.QueryRow(ctx, query, userID, subj).Scan(&n)
	return n, err
}

// ListLeaderboard: nama siswa, sekolah, nilai. Urutan: nilai tertinggi DESC, waktu tercepat ASC, nama ASC; belum mengerjakan = nama ASC di akhir.
// Nilai diambil dari database (attempts.score); semua attempt submitted termasuk (COALESCE score untuk urutan).
func (r *tryoutRegistrationRepo) ListLeaderboard(ctx context.Context, tryoutID string) ([]domain.LeaderboardEntry, error) {
	query := `
		WITH best_attempt AS (
			SELECT DISTINCT ON (a.user_id)
				a.user_id,
				a.score AS best_score,
				a.time_seconds_spent AS best_time_seconds
			FROM attempts a
			WHERE a.tryout_session_id = $1::uuid AND a.status = 'submitted'
			ORDER BY a.user_id, COALESCE(a.score, 0) DESC, a.time_seconds_spent ASC NULLS LAST
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
		ORDER BY COALESCE(b.best_score, 0) DESC, b.best_time_seconds ASC NULLS LAST, u.name ASC
	`
	rows, err := r.pool.Query(ctx, query, tryoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]domain.LeaderboardEntry, 0)
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

// ListLeaderboardRedisSyncRows mengikuti aturan "best attempt" yang sama dengan ListLeaderboard (nilai tertinggi, waktu tercepat).
func (r *tryoutRegistrationRepo) ListLeaderboardRedisSyncRows(ctx context.Context, tryoutID string) ([]LeaderboardRedisSyncRow, error) {
	const q = `
SELECT DISTINCT ON (a.user_id)
  a.user_id::text,
  a.score::float8
FROM attempts a
INNER JOIN tryout_registrations r
  ON r.user_id = a.user_id AND r.tryout_session_id = a.tryout_session_id
WHERE a.tryout_session_id = $1::uuid
  AND a.status = 'submitted'
  AND a.score IS NOT NULL
ORDER BY a.user_id, COALESCE(a.score, 0) DESC, a.time_seconds_spent ASC NULLS LAST
`
	rows, err := r.pool.Query(ctx, q, tryoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LeaderboardRedisSyncRow
	for rows.Next() {
		var uid string
		var score float64
		if err := rows.Scan(&uid, &score); err != nil {
			return nil, err
		}
		out = append(out, LeaderboardRedisSyncRow{UserID: uid, Score: score})
	}
	return out, rows.Err()
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
