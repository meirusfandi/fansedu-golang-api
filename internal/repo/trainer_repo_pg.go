package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var ErrStudentAlreadyLinked = errors.New("student already linked to this trainer")

type trainerRepo struct {
	pool *pgxpool.Pool
}

func NewTrainerRepo(pool *pgxpool.Pool) TrainerRepo {
	return &trainerRepo{pool: pool}
}

func (r *trainerRepo) GetOrCreateSlots(ctx context.Context, trainerID string) (int, error) {
	var paidSlots int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(paid_slots, 0) FROM trainer_slots WHERE trainer_id = $1::uuid
	`, trainerID).Scan(&paidSlots)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return paidSlots, nil
}

func (r *trainerRepo) AddSlots(ctx context.Context, trainerID string, quantity int) error {
	if quantity <= 0 {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO trainer_slots (trainer_id, paid_slots)
		VALUES ($1::uuid, $2)
		ON CONFLICT (trainer_id) DO UPDATE SET
			paid_slots = trainer_slots.paid_slots + $2,
			updated_at = NOW()
	`, trainerID, quantity)
	return err
}

func (r *trainerRepo) CountStudents(ctx context.Context, trainerID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM trainer_students WHERE trainer_id = $1::uuid
	`, trainerID).Scan(&n)
	return n, err
}

func (r *trainerRepo) ListStudents(ctx context.Context, trainerID string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.email, u.password_hash, u.name, u.role, u.avatar_url, u.school_id, u.subject_id, u.email_verified, u.email_verified_at, u.must_set_password, u.created_at, u.updated_at
		FROM trainer_students ts
		JOIN users u ON u.id = ts.student_id
		WHERE ts.trainer_id = $1::uuid
		ORDER BY ts.created_at ASC
	`, trainerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.User
	for rows.Next() {
		var u domain.User
		var avatarURL, schoolID, subjectID *string
		var emailVerifiedAt *time.Time
		var emailVerified, mustSetPassword bool
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &avatarURL, &schoolID, &subjectID, &emailVerified, &emailVerifiedAt, &mustSetPassword, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.AvatarURL = avatarURL
		u.SchoolID = schoolID
		u.SubjectID = subjectID
		u.EmailVerified = emailVerified
		u.EmailVerifiedAt = emailVerifiedAt
		u.MustSetPassword = mustSetPassword
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *trainerRepo) LinkStudent(ctx context.Context, trainerID, studentID string) error {
	cmd, err := r.pool.Exec(ctx, `
		INSERT INTO trainer_students (trainer_id, student_id) VALUES ($1::uuid, $2::uuid)
	`, trainerID, studentID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrStudentAlreadyLinked
		}
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrStudentAlreadyLinked
	}
	return nil
}

func (r *trainerRepo) ListTrainersByStudent(ctx context.Context, studentID string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.email, u.password_hash, u.name, u.role, u.avatar_url, u.school_id, u.subject_id,
		       u.email_verified, u.email_verified_at, u.must_set_password, u.created_at, u.updated_at
		FROM trainer_students ts
		JOIN users u ON u.id = ts.trainer_id
		WHERE ts.student_id = $1::uuid
		ORDER BY ts.created_at DESC
	`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		var avatarURL, schoolID, subjectID *string
		var emailVerifiedAt *time.Time
		var emailVerified, mustSetPassword bool
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &avatarURL, &schoolID, &subjectID,
			&emailVerified, &emailVerifiedAt, &mustSetPassword, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		u.AvatarURL = avatarURL
		u.SchoolID = schoolID
		u.SubjectID = subjectID
		u.EmailVerified = emailVerified
		u.EmailVerifiedAt = emailVerifiedAt
		u.MustSetPassword = mustSetPassword
		list = append(list, u)
	}
	return list, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}
