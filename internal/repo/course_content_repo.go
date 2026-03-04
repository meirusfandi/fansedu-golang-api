package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CourseContentRepo interface {
	Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
	GetByID(ctx context.Context, id string) (domain.CourseContent, error)
	ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error)
	Update(ctx context.Context, c domain.CourseContent) error
	Delete(ctx context.Context, id string) error
}

type courseContentRepo struct{ pool *pgxpool.Pool }

func NewCourseContentRepo(pool *pgxpool.Pool) CourseContentRepo {
	return &courseContentRepo{pool: pool}
}

func (r *courseContentRepo) Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO course_contents (id, course_id, title, description, sort_order, type, content)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6::course_content_type, $7)
		RETURNING created_at, updated_at
	`, id, c.CourseID, c.Title, c.Description, c.SortOrder, c.Type, c.Content).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.CourseContent{}, err
	}
	c.ID = id
	return c, nil
}

func (r *courseContentRepo) GetByID(ctx context.Context, id string) (domain.CourseContent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, course_id, title, description, sort_order, type, content, created_at, updated_at
		FROM course_contents WHERE id = $1::uuid
	`, id)
	var c domain.CourseContent
	err := row.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.SortOrder, &c.Type, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *courseContentRepo) ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, course_id, title, description, sort_order, type, content, created_at, updated_at
		FROM course_contents WHERE course_id = $1::uuid ORDER BY sort_order, created_at
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.CourseContent
	for rows.Next() {
		var c domain.CourseContent
		if err := rows.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.SortOrder, &c.Type, &c.Content, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *courseContentRepo) Update(ctx context.Context, c domain.CourseContent) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE course_contents SET title=$2, description=$3, sort_order=$4, type=$5::course_content_type, content=$6
		WHERE id = $1::uuid
	`, c.ID, c.Title, c.Description, c.SortOrder, c.Type, c.Content)
	return err
}

func (r *courseContentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM course_contents WHERE id = $1::uuid`, id)
	return err
}
