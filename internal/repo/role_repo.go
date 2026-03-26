package repo

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type RoleRepo interface {
	Create(ctx context.Context, e domain.Role) (domain.Role, error)
	GetByID(ctx context.Context, id string) (domain.Role, error)
	GetBySlug(ctx context.Context, slug string) (domain.Role, error)
	// GetByUserRoleCode finds a row whose effective code (COALESCE user_role_code, slug) matches users.role / JWT.
	GetByUserRoleCode(ctx context.Context, code string) (domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
	Update(ctx context.Context, e domain.Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepo struct{ pool *pgxpool.Pool }

func NewRoleRepo(pool *pgxpool.Pool) RoleRepo { return &roleRepo{pool: pool} }

func (r *roleRepo) Create(ctx context.Context, e domain.Role) (domain.Role, error) {
	id := uuid.New().String()
	code := strings.TrimSpace(e.UserRoleCode)
	if code == "" {
		code = strings.TrimSpace(e.Slug)
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO roles (id, name, slug, description, icon_url, user_role_code)
		VALUES ($1::uuid, $2, $3, $4, $5, NULLIF(trim($6), ''))
		RETURNING created_at, updated_at
	`, id, e.Name, e.Slug, e.Description, e.IconURL, code).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Role{}, err
	}
	e.ID = id
	e.UserRoleCode = code
	return e, nil
}

func (r *roleRepo) GetByID(ctx context.Context, id string) (domain.Role, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(NULLIF(trim(user_role_code), ''), slug), description, icon_url, created_at, updated_at
		FROM roles WHERE id = $1::uuid
	`, id)
	var e domain.Role
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.UserRoleCode, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *roleRepo) GetBySlug(ctx context.Context, slug string) (domain.Role, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(NULLIF(trim(user_role_code), ''), slug), description, icon_url, created_at, updated_at
		FROM roles WHERE lower(trim(slug)) = lower(trim($1))
	`, slug)
	var e domain.Role
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.UserRoleCode, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *roleRepo) GetByUserRoleCode(ctx context.Context, code string) (domain.Role, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, COALESCE(NULLIF(trim(user_role_code), ''), slug), description, icon_url, created_at, updated_at
		FROM roles
		WHERE lower(trim(COALESCE(NULLIF(trim(user_role_code), ''), slug))) = lower(trim($1))
		ORDER BY name
		LIMIT 1
	`, code)
	var e domain.Role
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.UserRoleCode, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *roleRepo) List(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, COALESCE(NULLIF(trim(user_role_code), ''), slug), description, icon_url, created_at, updated_at
		FROM roles ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Role
	for rows.Next() {
		var e domain.Role
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.UserRoleCode, &e.Description, &e.IconURL, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *roleRepo) Update(ctx context.Context, e domain.Role) error {
	code := strings.TrimSpace(e.UserRoleCode)
	if code == "" {
		code = strings.TrimSpace(e.Slug)
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE roles SET name=$2, slug=$3, description=$4, icon_url=$5, user_role_code=NULLIF(trim($6), '') WHERE id = $1::uuid
	`, e.ID, e.Name, e.Slug, e.Description, e.IconURL, code)
	return err
}

func (r *roleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE id = $1::uuid`, id)
	return err
}
