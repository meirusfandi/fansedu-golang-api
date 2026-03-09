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
	GetBySlug(ctx context.Context, slug string) (domain.Course, error)
	List(ctx context.Context) ([]domain.Course, error)
	ListBySubjectID(ctx context.Context, subjectID *string) ([]domain.Course, error)
	ListByCreatedBy(ctx context.Context, createdBy string) ([]domain.Course, error)
	Update(ctx context.Context, c domain.Course) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}

type courseRepo struct{ pool *pgxpool.Pool }

func NewCourseRepo(pool *pgxpool.Pool) CourseRepo { return &courseRepo{pool: pool} }

func (r *courseRepo) Create(ctx context.Context, c domain.Course) (domain.Course, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO courses (id, title, slug, description, price_cents, thumbnail, subject_id, created_by)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid, $8::uuid)
		RETURNING created_at, updated_at
	`, id, c.Title, c.Slug, c.Description, c.PriceCents, c.Thumbnail, c.SubjectID, c.CreatedBy).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return domain.Course{}, err
	}
	c.ID = id
	return c, nil
}

func (r *courseRepo) GetByID(ctx context.Context, id string) (domain.Course, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
		FROM courses WHERE id = $1::uuid
	`, id)
	var c domain.Course
	err := row.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *courseRepo) GetBySlug(ctx context.Context, slug string) (domain.Course, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
		FROM courses WHERE slug = $1
	`, slug)
	var c domain.Course
	err := row.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *courseRepo) List(ctx context.Context) ([]domain.Course, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
		FROM courses ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *courseRepo) ListBySubjectID(ctx context.Context, subjectID *string) ([]domain.Course, error) {
	if subjectID == nil || *subjectID == "" {
		rows, err := r.pool.Query(ctx, `
			SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
			FROM courses WHERE subject_id IS NULL ORDER BY created_at DESC
		`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var list []domain.Course
		for rows.Next() {
			var c domain.Course
			if err := rows.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
				return nil, err
			}
			list = append(list, c)
		}
		return list, rows.Err()
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
		FROM courses WHERE subject_id = $1::uuid OR subject_id IS NULL ORDER BY created_at DESC
	`, *subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *courseRepo) ListByCreatedBy(ctx context.Context, createdBy string) ([]domain.Course, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, slug, description, price_cents, thumbnail, subject_id, created_by, created_at, updated_at
		FROM courses WHERE created_by = $1::uuid ORDER BY created_at DESC
	`, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Course
	for rows.Next() {
		var c domain.Course
		if err := rows.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.PriceCents, &c.Thumbnail, &c.SubjectID, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
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
	_, err := r.pool.Exec(ctx, `
		UPDATE courses SET title=$2, slug=$3, description=$4, price_cents=$5, thumbnail=$6, subject_id=$7::uuid WHERE id = $1::uuid
	`, c.ID, c.Title, c.Slug, c.Description, c.PriceCents, c.Thumbnail, c.SubjectID)
	return err
}

func (r *courseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM courses WHERE id = $1::uuid`, id)
	return err
}
