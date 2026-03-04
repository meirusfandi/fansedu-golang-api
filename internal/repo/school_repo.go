package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type SchoolRepo interface {
	Create(ctx context.Context, e domain.School) (domain.School, error)
	GetByID(ctx context.Context, id string) (domain.School, error)
	GetBySlug(ctx context.Context, slug string) (domain.School, error)
	List(ctx context.Context) ([]domain.School, error)
	Update(ctx context.Context, e domain.School) error
	Delete(ctx context.Context, id string) error
}

type schoolRepo struct{ pool *pgxpool.Pool }

func NewSchoolRepo(pool *pgxpool.Pool) SchoolRepo { return &schoolRepo{pool: pool} }

func (r *schoolRepo) Create(ctx context.Context, e domain.School) (domain.School, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO schools (id, name, slug, description, address, logo_url)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, e.Name, e.Slug, e.Description, e.Address, e.LogoURL).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.School{}, err
	}
	e.ID = id
	return e, nil
}

func (r *schoolRepo) GetByID(ctx context.Context, id string) (domain.School, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, address, logo_url, created_at, updated_at
		FROM schools WHERE id = $1::uuid
	`, id)
	var e domain.School
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.Address, &e.LogoURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *schoolRepo) GetBySlug(ctx context.Context, slug string) (domain.School, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, address, logo_url, created_at, updated_at
		FROM schools WHERE slug = $1
	`, slug)
	var e domain.School
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.Address, &e.LogoURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *schoolRepo) List(ctx context.Context) ([]domain.School, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, description, address, logo_url, created_at, updated_at
		FROM schools ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.School
	for rows.Next() {
		var e domain.School
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.Address, &e.LogoURL, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *schoolRepo) Update(ctx context.Context, e domain.School) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE schools SET name=$2, slug=$3, description=$4, address=$5, logo_url=$6 WHERE id = $1::uuid
	`, e.ID, e.Name, e.Slug, e.Description, e.Address, e.LogoURL)
	return err
}

func (r *schoolRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM schools WHERE id = $1::uuid`, id)
	return err
}
