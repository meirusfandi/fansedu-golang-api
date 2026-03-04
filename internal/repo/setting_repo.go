package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type SettingRepo interface {
	Create(ctx context.Context, e domain.Setting) (domain.Setting, error)
	GetByID(ctx context.Context, id string) (domain.Setting, error)
	GetByKey(ctx context.Context, key string) (domain.Setting, error)
	List(ctx context.Context) ([]domain.Setting, error)
	Update(ctx context.Context, e domain.Setting) error
	Delete(ctx context.Context, id string) error
}

type settingRepo struct{ pool *pgxpool.Pool }

func NewSettingRepo(pool *pgxpool.Pool) SettingRepo { return &settingRepo{pool: pool} }

func (r *settingRepo) Create(ctx context.Context, e domain.Setting) (domain.Setting, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO settings (id, key, slug, value, value_json, description)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, e.Key, e.Slug, e.Value, e.ValueJSON, e.Description).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Setting{}, err
	}
	e.ID = id
	return e, nil
}

func (r *settingRepo) GetByID(ctx context.Context, id string) (domain.Setting, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, key, slug, value, value_json, description, created_at, updated_at
		FROM settings WHERE id = $1::uuid
	`, id)
	var e domain.Setting
	err := row.Scan(&e.ID, &e.Key, &e.Slug, &e.Value, &e.ValueJSON, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *settingRepo) GetByKey(ctx context.Context, key string) (domain.Setting, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, key, slug, value, value_json, description, created_at, updated_at
		FROM settings WHERE key = $1
	`, key)
	var e domain.Setting
	err := row.Scan(&e.ID, &e.Key, &e.Slug, &e.Value, &e.ValueJSON, &e.Description, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *settingRepo) List(ctx context.Context) ([]domain.Setting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, key, slug, value, value_json, description, created_at, updated_at
		FROM settings ORDER BY key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Setting
	for rows.Next() {
		var e domain.Setting
		if err := rows.Scan(&e.ID, &e.Key, &e.Slug, &e.Value, &e.ValueJSON, &e.Description, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *settingRepo) Update(ctx context.Context, e domain.Setting) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE settings SET key=$2, slug=$3, value=$4, value_json=$5, description=$6 WHERE id = $1::uuid
	`, e.ID, e.Key, e.Slug, e.Value, e.ValueJSON, e.Description)
	return err
}

func (r *settingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM settings WHERE id = $1::uuid`, id)
	return err
}
