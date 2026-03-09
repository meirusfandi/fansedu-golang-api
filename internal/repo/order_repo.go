package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type OrderRepo interface {
	Create(ctx context.Context, o domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id, status string) error
}

type orderRepo struct{ pool *pgxpool.Pool }

func NewOrderRepo(pool *pgxpool.Pool) OrderRepo { return &orderRepo{pool: pool} }

func (r *orderRepo) Create(ctx context.Context, o domain.Order) (domain.Order, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO orders (id, user_id, status, total_price_cents, payment_method, payment_reference)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, id, o.UserID, o.Status, o.TotalPriceCents, o.PaymentMethod, o.PaymentReference).Scan(&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}
	o.ID = id
	return o, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (domain.Order, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, total_price_cents, payment_method, payment_reference, created_at, updated_at
		FROM orders WHERE id = $1::uuid
	`, id)
	var o domain.Order
	err := row.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPriceCents, &o.PaymentMethod, &o.PaymentReference, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *orderRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status, total_price_cents, payment_method, payment_reference, created_at, updated_at
		FROM orders WHERE user_id = $1::uuid ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPriceCents, &o.PaymentMethod, &o.PaymentReference, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1::uuid`, id, status)
	return err
}
