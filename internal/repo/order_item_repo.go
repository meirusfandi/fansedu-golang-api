package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type OrderItemRepo interface {
	Create(ctx context.Context, oi domain.OrderItem) (domain.OrderItem, error)
	ListByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error)
}

type orderItemRepo struct{ pool *pgxpool.Pool }

func NewOrderItemRepo(pool *pgxpool.Pool) OrderItemRepo { return &orderItemRepo{pool: pool} }

func (r *orderItemRepo) Create(ctx context.Context, oi domain.OrderItem) (domain.OrderItem, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO order_items (id, order_id, course_id, price)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4)
		RETURNING created_at
	`, id, oi.OrderID, oi.CourseID, oi.Price).Scan(&oi.CreatedAt)
	if err != nil {
		return domain.OrderItem{}, err
	}
	oi.ID = id
	return oi, nil
}

func (r *orderItemRepo) ListByOrderID(ctx context.Context, orderID string) ([]domain.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, order_id, course_id, price, created_at
		FROM order_items WHERE order_id = $1::uuid ORDER BY created_at
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.OrderItem
	for rows.Next() {
		var oi domain.OrderItem
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.CourseID, &oi.Price, &oi.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, oi)
	}
	return list, rows.Err()
}
