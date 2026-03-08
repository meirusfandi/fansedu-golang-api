package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type NotificationRepo interface {
	Create(ctx context.Context, n domain.Notification) (domain.Notification, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]domain.Notification, error)
	MarkRead(ctx context.Context, id, userID string) error
}

type notificationRepo struct{ pool *pgxpool.Pool }

func NewNotificationRepo(pool *pgxpool.Pool) NotificationRepo {
	return &notificationRepo{pool: pool}
}

func (r *notificationRepo) Create(ctx context.Context, n domain.Notification) (domain.Notification, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (id, user_id, title, body, type)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5)
		RETURNING created_at
	`, id, n.UserID, n.Title, n.Body, n.Type).Scan(&n.CreatedAt)
	if err != nil {
		return domain.Notification{}, err
	}
	n.ID = id
	return n, nil
}

func (r *notificationRepo) ListByUserID(ctx context.Context, userID string, limit int) ([]domain.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, title, body, type, read_at, created_at
		FROM notifications WHERE user_id = $1::uuid ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, n)
	}
	return list, rows.Err()
}

func (r *notificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications SET read_at = NOW() WHERE id = $1::uuid AND user_id = $2::uuid
	`, id, userID)
	return err
}
