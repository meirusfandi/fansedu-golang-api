package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CourseMessageRepo interface {
	Create(ctx context.Context, m domain.CourseMessage) (domain.CourseMessage, error)
	ListByCourseID(ctx context.Context, courseID string, limit int) ([]domain.CourseMessage, error)
}

type courseMessageRepo struct{ pool *pgxpool.Pool }

func NewCourseMessageRepo(pool *pgxpool.Pool) CourseMessageRepo {
	return &courseMessageRepo{pool: pool}
}

func (r *courseMessageRepo) Create(ctx context.Context, m domain.CourseMessage) (domain.CourseMessage, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_messages (id, course_id, user_id, message)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4)
		RETURNING created_at
	`, id, m.CourseID, m.UserID, m.Message).Scan(&m.CreatedAt)
	if err != nil {
		return domain.CourseMessage{}, err
	}
	m.ID = id
	return m, nil
}

func (r *courseMessageRepo) ListByCourseID(ctx context.Context, courseID string, limit int) ([]domain.CourseMessage, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, course_id, user_id, message, created_at
		FROM course_messages WHERE course_id = $1::uuid ORDER BY created_at DESC LIMIT $2
	`, courseID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseMessage
	for rows.Next() {
		var m domain.CourseMessage
		if err := rows.Scan(&m.ID, &m.CourseID, &m.UserID, &m.Message, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}
