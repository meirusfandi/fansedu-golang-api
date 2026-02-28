package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type EnrollmentRepo interface {
	Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.CourseEnrollment, error)
	ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	Update(ctx context.Context, e domain.CourseEnrollment) error
}

type enrollmentRepo struct{ pool *pgxpool.Pool }

func NewEnrollmentRepo(pool *pgxpool.Pool) EnrollmentRepo { return &enrollmentRepo{pool: pool} }

func (r *enrollmentRepo) Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_enrollments (id, user_id, course_id, status)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::enrollment_status)
		RETURNING id, enrolled_at, created_at
	`, id, e.UserID, e.CourseID, e.Status).Scan(&e.ID, &e.EnrolledAt, &e.CreatedAt)
	if err != nil {
		return domain.CourseEnrollment{}, err
	}
	e.ID = id
	return e, nil
}

func (r *enrollmentRepo) GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, course_id, status, enrolled_at, completed_at, created_at
		FROM course_enrollments WHERE user_id = $1::uuid AND course_id = $2::uuid
	`, userID, courseID)
	var e domain.CourseEnrollment
	err := row.Scan(&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.EnrolledAt, &e.CompletedAt, &e.CreatedAt)
	return e, err
}

func (r *enrollmentRepo) ListByUserID(ctx context.Context, userID string) ([]domain.CourseEnrollment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, course_id, status, enrolled_at, completed_at, created_at
		FROM course_enrollments WHERE user_id = $1::uuid ORDER BY enrolled_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseEnrollment
	for rows.Next() {
		var e domain.CourseEnrollment
		if err := rows.Scan(&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.EnrolledAt, &e.CompletedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *enrollmentRepo) ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, course_id, status, enrolled_at, completed_at, created_at
		FROM course_enrollments WHERE course_id = $1::uuid ORDER BY enrolled_at DESC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseEnrollment
	for rows.Next() {
		var e domain.CourseEnrollment
		if err := rows.Scan(&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.EnrolledAt, &e.CompletedAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *enrollmentRepo) Update(ctx context.Context, e domain.CourseEnrollment) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE course_enrollments SET status=$2::enrollment_status, completed_at=$3 WHERE id = $1::uuid
	`, e.ID, e.Status, e.CompletedAt)
	return err
}
