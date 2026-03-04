package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type RoleRepo interface {
	Create(ctx context.Context, e domain.Role) (domain.Role, error)
	GetByID(ctx context.Context, id string) (domain.Role, error)
	GetBySlug(ctx context.Context, slug string) (domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
	Update(ctx context.Context, e domain.Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepo struct{ pool *pgxpool.Pool }

func NewRoleRepo(pool *pgxpool.Pool) RoleRepo { return &roleRepo{pool: pool} }

func (r *roleRepo) Create(ctx context.Context, e domain.Role) (domain.Role, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO roles (id, name, slug, description, icon_url)
		VALUES ($1::uuid, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`, id, e.Name, e.Slug, e.Description, e.IconURL).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Role{}, err
	}
	e.ID = id
	return e, nil
}

func (r *roleRepo) GetByID(ctx context.Context, id string) (domain.Role, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, icon_url, created_at, updated_at
		FROM roles WHERE id = $1::uuid
	`, id)
	var e domain.Role
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *roleRepo) GetBySlug(ctx context.Context, slug string) (domain.Role, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, icon_url, created_at, updated_at
		FROM roles WHERE slug = $1
	`, slug)
	var e domain.Role
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *roleRepo) List(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, description, icon_url, created_at, updated_at
		FROM roles ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Role
	for rows.Next() {
		var e domain.Role
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *roleRepo) Update(ctx context.Context, e domain.Role) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE roles SET name=$2, slug=$3, description=$4, icon_url=$5 WHERE id = $1::uuid
	`, e.ID, e.Name, e.Slug, e.Description, e.IconURL)
	return err
}

func (r *roleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1::uuid`, id)
	return err
}
