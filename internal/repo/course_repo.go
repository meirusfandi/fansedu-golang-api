package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CourseRepo interface {
	Create(ctx context.Context, c domain.Course) (domain.Course, error)
	GetByID(ctx context.Context, id string) (domain.Course, error)
	List(ctx context.Context) ([]domain.Course, error)
	Update(ctx context.Context, c domain.Course) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}

type courseRepo struct{ pool *pgxpool.Pool }

func NewCourseRepo(pool *pgxpool.Pool) CourseRepo { return &courseRepo{pool: pool} }

func (r *courseRepo) Create(ctx context.Context, c domain.Course) (domain.Course, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO courses (id, title, description, created_by)
		VALUES ($1::uuid, $2, $3, $4::uuid)
		RETURNING created_at, updated_at
	`, id, c.Title, c.Description, c.CreatedBy).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Course{}, err
	}
	c.ID = id
	return c, nil
}

func (r *courseRepo) GetByID(ctx context.Context, id string) (domain.Course, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, description, created_by, created_at, updated_at
		FROM courses WHERE id = $1::uuid
	`, id)
	var c domain.Course
	err := row.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *courseRepo) List(ctx context.Context) ([]domain.Course, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, description, created_by, created_at, updated_at
		FROM courses ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *courseRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM courses`).Scan(&n)
	return n, err
}

func (r *courseRepo) Update(ctx context.Context, c domain.Course) error {
	_, err := r.pool.Exec(ctx, `UPDATE courses SET title=$2, description=$3 WHERE id = $1::uuid`, c.ID, c.Title, c.Description)
	return err
}

func (r *courseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM courses WHERE id = $1::uuid`, id)
	return err
}
