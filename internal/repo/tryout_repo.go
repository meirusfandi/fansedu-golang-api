package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type TryoutRepo interface {
	Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
	List(ctx context.Context) ([]domain.TryoutSession, error)
	ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
	ListOpenForStudent(ctx context.Context, now time.Time, subjectID *string) ([]domain.TryoutSession, error)
	ListForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error)
	Update(ctx context.Context, t domain.TryoutSession) error
	Delete(ctx context.Context, id string) error
}

type tryoutRepo struct{ pool *pgxpool.Pool }

func NewTryoutRepo(pool *pgxpool.Pool) TryoutRepo { return &tryoutRepo{pool: pool} }

func normalizeTryoutGradingMode(s string) string {
	switch s {
	case domain.TryoutGradingModeManual:
		return domain.TryoutGradingModeManual
	default:
		return domain.TryoutGradingModeAuto
	}
}

func (r *tryoutRepo) Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error) {
	var id string
	gm := normalizeTryoutGradingMode(t.GradingMode)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tryout_sessions (title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode, created_by)
		VALUES ($1, $2, $3, $4, $5, $6::tryout_level, $7, $8, $9::uuid, $10::uuid, $11, $12, $13, $14::tryout_status, $15::tryout_grading_mode, $16::uuid)
		RETURNING id
	`, t.Title, t.ShortTitle, t.Description, t.DurationMinutes, t.QuestionsCount, t.Level, t.Subject, t.SchoolLevel, t.SubjectID, t.LevelID, t.OpensAt, t.ClosesAt, t.MaxParticipants, t.Status, gm, t.CreatedBy).Scan(&id)
	if err != nil {
		return domain.TryoutSession{}, err
	}
	t.ID = id
	t.GradingMode = gm
	return t, nil
}

func (r *tryoutRepo) GetByID(ctx context.Context, id string) (domain.TryoutSession, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode::text, created_by, created_at, updated_at
		FROM tryout_sessions WHERE id = $1::uuid
	`, id)
	var t domain.TryoutSession
	var subjectID *string
	err := row.Scan(&t.ID, &t.Title, &t.ShortTitle, &t.Description, &t.DurationMinutes, &t.QuestionsCount, &t.Level, &t.Subject, &t.SchoolLevel, &subjectID, &t.LevelID, &t.OpensAt, &t.ClosesAt, &t.MaxParticipants, &t.Status, &t.GradingMode, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	t.SubjectID = subjectID
	return t, err
}

func (r *tryoutRepo) List(ctx context.Context) ([]domain.TryoutSession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode::text, created_by, created_at, updated_at
		FROM tryout_sessions ORDER BY opens_at DESC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TryoutSession
	for rows.Next() {
		var t domain.TryoutSession
		var subjectID *string
		if err := rows.Scan(&t.ID, &t.Title, &t.ShortTitle, &t.Description, &t.DurationMinutes, &t.QuestionsCount, &t.Level, &t.Subject, &t.SchoolLevel, &subjectID, &t.LevelID, &t.OpensAt, &t.ClosesAt, &t.MaxParticipants, &t.Status, &t.GradingMode, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.SubjectID = subjectID
		list = append(list, t)
	}
	return list, rows.Err()
}

// ListOpen: status open dan belum lewat closes_at (masih dalam periode pendaftaran/penyelenggaraan).
func (r *tryoutRepo) ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode::text, created_by, created_at, updated_at
		FROM tryout_sessions
		WHERE status = 'open' AND closes_at >= $1
		ORDER BY opens_at NULLS LAST, created_at DESC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TryoutSession
	for rows.Next() {
		var t domain.TryoutSession
		var subjectID *string
		if err := rows.Scan(&t.ID, &t.Title, &t.ShortTitle, &t.Description, &t.DurationMinutes, &t.QuestionsCount, &t.Level, &t.Subject, &t.SchoolLevel, &subjectID, &t.LevelID, &t.OpensAt, &t.ClosesAt, &t.MaxParticipants, &t.Status, &t.GradingMode, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.SubjectID = subjectID
		list = append(list, t)
	}
	return list, rows.Err()
}

// ListOpenForStudent: status open, closes_at belum lewat, + filter bidang siswa.
func (r *tryoutRepo) ListOpenForStudent(ctx context.Context, now time.Time, subjectID *string) ([]domain.TryoutSession, error) {
	query := `
		SELECT id, title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode::text, created_by, created_at, updated_at
		FROM tryout_sessions
		WHERE status = 'open' AND closes_at >= $1
		AND (subject_id IS NULL OR ($2::text IS NOT NULL AND subject_id = $2::uuid))
		ORDER BY opens_at NULLS LAST, created_at DESC
	`
	var subj interface{}
	if subjectID != nil && *subjectID != "" {
		subj = *subjectID
	}
	rows, err := r.pool.Query(ctx, query, now, subj)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TryoutSession
	for rows.Next() {
		var t domain.TryoutSession
		var sid *string
		if err := rows.Scan(&t.ID, &t.Title, &t.ShortTitle, &t.Description, &t.DurationMinutes, &t.QuestionsCount, &t.Level, &t.Subject, &t.SchoolLevel, &sid, &t.LevelID, &t.OpensAt, &t.ClosesAt, &t.MaxParticipants, &t.Status, &t.GradingMode, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.SubjectID = sid
		list = append(list, t)
	}
	return list, rows.Err()
}

// ListForStudent returns all tryouts for the student's subject (subject_id IS NULL or = subjectID), excluding draft.
// Status open/closed included; frontend can separate by status or opens_at/closes_at.
func (r *tryoutRepo) ListForStudent(ctx context.Context, subjectID *string) ([]domain.TryoutSession, error) {
	query := `
		SELECT id, title, short_title, description, duration_minutes, questions_count, level, subject, school_level, subject_id, level_id, opens_at, closes_at, max_participants, status, grading_mode::text, created_by, created_at, updated_at
		FROM tryout_sessions
		WHERE status != 'draft'
		AND (subject_id IS NULL OR ($1::text IS NOT NULL AND subject_id = $1::uuid))
		ORDER BY opens_at DESC, created_at DESC
	`
	var subj interface{}
	if subjectID != nil && *subjectID != "" {
		subj = *subjectID
	}
	rows, err := r.pool.Query(ctx, query, subj)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.TryoutSession
	for rows.Next() {
		var t domain.TryoutSession
		var sid *string
		if err := rows.Scan(&t.ID, &t.Title, &t.ShortTitle, &t.Description, &t.DurationMinutes, &t.QuestionsCount, &t.Level, &t.Subject, &t.SchoolLevel, &sid, &t.LevelID, &t.OpensAt, &t.ClosesAt, &t.MaxParticipants, &t.Status, &t.GradingMode, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.SubjectID = sid
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *tryoutRepo) Update(ctx context.Context, t domain.TryoutSession) error {
	gm := normalizeTryoutGradingMode(t.GradingMode)
	_, err := r.pool.Exec(ctx, `
		UPDATE tryout_sessions SET title=$2, short_title=$3, description=$4, duration_minutes=$5, questions_count=$6, level=$7::tryout_level, subject=$8, school_level=$9, subject_id=$10::uuid, level_id=$11::uuid, opens_at=$12, closes_at=$13, max_participants=$14, status=$15::tryout_status, grading_mode=$16::tryout_grading_mode
		WHERE id = $1::uuid
	`, t.ID, t.Title, t.ShortTitle, t.Description, t.DurationMinutes, t.QuestionsCount, t.Level, t.Subject, t.SchoolLevel, t.SubjectID, t.LevelID, t.OpensAt, t.ClosesAt, t.MaxParticipants, t.Status, gm)
	return err
}

func (r *tryoutRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tryout_sessions WHERE id = $1::uuid`, id)
	return err
}
