package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type EventRepo interface {
	Create(ctx context.Context, e domain.Event) (domain.Event, error)
	GetByID(ctx context.Context, id string) (domain.Event, error)
	GetBySlug(ctx context.Context, slug string) (domain.Event, error)
	List(ctx context.Context) ([]domain.Event, error)
	Update(ctx context.Context, e domain.Event) error
	Delete(ctx context.Context, id string) error
}

type eventRepo struct{ pool *pgxpool.Pool }

func NewEventRepo(pool *pgxpool.Pool) EventRepo { return &eventRepo{pool: pool} }

func (r *eventRepo) Create(ctx context.Context, e domain.Event) (domain.Event, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO events (id, title, slug, description, start_at, end_at, thumbnail_url, status)
		VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`, id, e.Title, e.Slug, e.Description, e.StartAt, e.EndAt, e.ThumbnailURL, e.Status).Scan(&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Event{}, err
	}
	e.ID = id
	return e, nil
}

func (r *eventRepo) GetByID(ctx context.Context, id string) (domain.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, slug, description, start_at, end_at, thumbnail_url, status, created_at, updated_at
		FROM events WHERE id = $1::uuid
	`, id)
	var e domain.Event
	err := row.Scan(&e.ID, &e.Title, &e.Slug, &e.Description, &e.StartAt, &e.EndAt, &e.ThumbnailURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *eventRepo) GetBySlug(ctx context.Context, slug string) (domain.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, title, slug, description, start_at, end_at, thumbnail_url, status, created_at, updated_at
		FROM events WHERE slug = $1
	`, slug)
	var e domain.Event
	err := row.Scan(&e.ID, &e.Title, &e.Slug, &e.Description, &e.StartAt, &e.EndAt, &e.ThumbnailURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *eventRepo) List(ctx context.Context) ([]domain.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, slug, description, start_at, end_at, thumbnail_url, status, created_at, updated_at
		FROM events ORDER BY start_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Slug, &e.Description, &e.StartAt, &e.EndAt, &e.ThumbnailURL, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *eventRepo) Update(ctx context.Context, e domain.Event) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE events SET title=$2, slug=$3, description=$4, start_at=$5, end_at=$6, thumbnail_url=$7, status=$8 WHERE id = $1::uuid
	`, e.ID, e.Title, e.Slug, e.Description, e.StartAt, e.EndAt, e.ThumbnailURL, e.Status)
	return err
}

func (r *eventRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id = $1::uuid`, id)
	return err
}
