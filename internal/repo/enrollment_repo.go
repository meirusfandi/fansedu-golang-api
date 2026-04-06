package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type EnrollmentRepo interface {
	Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
	UpdateEnrolledAt(ctx context.Context, enrollmentID string, enrolledAt time.Time) error
	GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.CourseEnrollment, error)
	ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	ListCoursesByUserWithFilters(ctx context.Context, userID, search, progressStatus string, page, limit int) ([]domain.StudentCourseRow, int, error)
	Update(ctx context.Context, e domain.CourseEnrollment) error
	Count(ctx context.Context) (int, error)
	CountEnrolledInMonth(ctx context.Context, year, month int) (int, error)
}

type enrollmentRepo struct{ pool *pgxpool.Pool }

func NewEnrollmentRepo(pool *pgxpool.Pool) EnrollmentRepo { return &enrollmentRepo{pool: pool} }

func (r *enrollmentRepo) Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error) {
	id := uuid.New().String()
	var enrolledArg interface{}
	if !e.EnrolledAt.IsZero() {
		enrolledArg = e.EnrolledAt
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_enrollments (id, user_id, course_id, status, enrolled_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::enrollment_status, COALESCE($5::timestamptz, NOW()))
		RETURNING id, enrolled_at, created_at
	`, id, e.UserID, e.CourseID, e.Status, enrolledArg).Scan(&e.ID, &e.EnrolledAt, &e.CreatedAt)
	if err != nil {
		return domain.CourseEnrollment{}, err
	}
	e.ID = id
	return e, nil
}

func (r *enrollmentRepo) UpdateEnrolledAt(ctx context.Context, enrollmentID string, enrolledAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE course_enrollments SET enrolled_at = $2 WHERE id = $1::uuid
	`, enrollmentID, enrolledAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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

func (r *enrollmentRepo) ListCoursesByUserWithFilters(ctx context.Context, userID, search, progressStatus string, page, limit int) ([]domain.StudentCourseRow, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT
			e.id::text AS enrollment_id,
			c.id::text AS course_id,
			c.title,
			COALESCE(c.slug, '') AS course_slug,
			COALESCE(c.thumbnail, '') AS course_thumbnail,
			e.status AS enrollment_status,
			e.enrolled_at,
			COUNT(*) OVER()::int AS total
		FROM course_enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.user_id = $1::uuid
		  AND (
			$2 = '' OR
			c.title ILIKE '%' || $2 || '%' OR
			COALESCE(c.slug, '') ILIKE '%' || $2 || '%'
		  )
		  AND (
			$3 = '' OR
			($3 = 'in-progress' AND e.status = 'in_progress') OR
			($3 = 'completed' AND e.status = 'completed')
		  )
		ORDER BY e.enrolled_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, userID, search, progressStatus, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.StudentCourseRow, 0, limit)
	total := 0
	for rows.Next() {
		var row domain.StudentCourseRow
		if err := rows.Scan(&row.EnrollmentID, &row.CourseID, &row.CourseTitle, &row.CourseSlug, &row.CourseThumbnail, &row.EnrollmentStatus, &row.EnrolledAt, &total); err != nil {
			return nil, 0, err
		}
		out = append(out, row)
	}
	return out, total, rows.Err()
}

func (r *enrollmentRepo) Update(ctx context.Context, e domain.CourseEnrollment) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE course_enrollments SET status=$2::enrollment_status, completed_at=$3 WHERE id = $1::uuid
	`, e.ID, e.Status, e.CompletedAt)
	return err
}

func (r *enrollmentRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM course_enrollments`).Scan(&n)
	return n, err
}

func (r *enrollmentRepo) CountEnrolledInMonth(ctx context.Context, year, month int) (int, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM course_enrollments WHERE enrolled_at >= $1 AND enrolled_at < $2`, start, end).Scan(&n)
	return n, err
}
