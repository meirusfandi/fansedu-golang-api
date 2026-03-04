package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type SubjectRepo interface {
	Create(ctx context.Context, e domain.Subject) (domain.Subject, error)
	GetByID(ctx context.Context, id string) (domain.Subject, error)
	GetBySlug(ctx context.Context, slug string) (domain.Subject, error)
	List(ctx context.Context) ([]domain.Subject, error)
	Update(ctx context.Context, e domain.Subject) error
	Delete(ctx context.Context, id string) error
}

type subjectRepo struct{ pool *pgxpool.Pool }

func NewSubjectRepo(pool *pgxpool.Pool) SubjectRepo { return &subjectRepo{pool: pool} }

func (r *subjectRepo) Create(ctx context.Context, e domain.Subject) (domain.Subject, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO subjects (id, name, slug, description, icon_url, sort_order)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, e.Name, e.Slug, e.Description, e.IconURL, e.SortOrder).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Subject{}, err
	}
	e.ID = id
	return e, nil
}

func (r *subjectRepo) GetByID(ctx context.Context, id string) (domain.Subject, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, icon_url, sort_order, created_at, updated_at
		FROM subjects WHERE id = $1::uuid
	`, id)
	var e domain.Subject
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.SortOrder, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *subjectRepo) GetBySlug(ctx context.Context, slug string) (domain.Subject, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, icon_url, sort_order, created_at, updated_at
		FROM subjects WHERE slug = $1
	`, slug)
	var e domain.Subject
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.SortOrder, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *subjectRepo) List(ctx context.Context) ([]domain.Subject, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, description, icon_url, sort_order, created_at, updated_at
		FROM subjects ORDER BY sort_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Subject
	for rows.Next() {
		var e domain.Subject
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.SortOrder, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *subjectRepo) Update(ctx context.Context, e domain.Subject) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE subjects SET name=$2, slug=$3, description=$4, icon_url=$5, sort_order=$6 WHERE id = $1::uuid
	`, e.ID, e.Name, e.Slug, e.Description, e.IconURL, e.SortOrder)
	return err
}

func (r *subjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM subjects WHERE id = $1::uuid`, id)
	return err
}
