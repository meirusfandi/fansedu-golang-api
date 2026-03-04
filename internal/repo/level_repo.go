package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type LevelRepo interface {
	Create(ctx context.Context, e domain.Level) (domain.Level, error)
	GetByID(ctx context.Context, id string) (domain.Level, error)
	GetBySlug(ctx context.Context, slug string) (domain.Level, error)
	List(ctx context.Context) ([]domain.Level, error)
	Update(ctx context.Context, e domain.Level) error
	Delete(ctx context.Context, id string) error
	ListSubjectIDsByLevel(ctx context.Context, levelID string) ([]string, error)
}

type levelRepo struct{ pool *pgxpool.Pool }

func NewLevelRepo(pool *pgxpool.Pool) LevelRepo { return &levelRepo{pool: pool} }

func (r *levelRepo) Create(ctx context.Context, e domain.Level) (domain.Level, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO levels (id, name, slug, description, sort_order, icon_url)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, e.Name, e.Slug, e.Description, e.SortOrder, e.IconURL).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Level{}, err
	}
	e.ID = id
	return e, nil
}

func (r *levelRepo) GetByID(ctx context.Context, id string) (domain.Level, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, sort_order, icon_url, created_at, updated_at
		FROM levels WHERE id = $1::uuid
	`, id)
	var e domain.Level
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.SortOrder, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *levelRepo) GetBySlug(ctx context.Context, slug string) (domain.Level, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, description, sort_order, icon_url, created_at, updated_at
		FROM levels WHERE slug = $1
	`, slug)
	var e domain.Level
	err := row.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.SortOrder, &e.IconURL, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *levelRepo) List(ctx context.Context) ([]domain.Level, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, description, sort_order, icon_url, created_at, updated_at
		FROM levels ORDER BY sort_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Level
	for rows.Next() {
		var e domain.Level
		if err := rows.Scan(&e.ID, &e.Name, &e.Slug, &e.Description, &e.SortOrder, &e.IconURL, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *levelRepo) Update(ctx context.Context, e domain.Level) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE levels SET name=$2, slug=$3, description=$4, sort_order=$5, icon_url=$6 WHERE id = $1::uuid
	`, e.ID, e.Name, e.Slug, e.Description, e.SortOrder, e.IconURL)
	return err
}

func (r *levelRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM levels WHERE id = $1::uuid`, id)
	return err
}

func (r *levelRepo) ListSubjectIDsByLevel(ctx context.Context, levelID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT subject_id FROM subject_levels WHERE level_id = $1::uuid ORDER BY sort_order
	`, levelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
