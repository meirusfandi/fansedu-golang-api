package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type PaymentRepo interface {
	Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
	List(ctx context.Context, limit int) ([]domain.Payment, error)
	CountPaidInMonth(ctx context.Context, year, month int) (int, error)
	TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
}

type paymentRepo struct{ pool *pgxpool.Pool }

func NewPaymentRepo(pool *pgxpool.Pool) PaymentRepo {
	return &paymentRepo{pool: pool}
}

func (r *paymentRepo) Create(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	id := uuid.New().String()
	var refArg interface{}
	if p.ReferenceID != nil {
		if u, err := uuid.Parse(*p.ReferenceID); err == nil {
			refArg = u
		}
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (id, user_id, amount_cents, currency, status, type, reference_id, description, paid_at)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5::payment_status, $6::payment_type, $7, $8, $9)
		RETURNING created_at, updated_at
	`, id, p.UserID, p.AmountCents, p.Currency, p.Status, p.Type, refArg, p.Description, p.PaidAt).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Payment{}, err
	}
	p.ID = id
	return p, nil
}

func (r *paymentRepo) List(ctx context.Context, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount_cents, currency, status, type, reference_id, description, paid_at, created_at, updated_at
		FROM payments ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var refID pgtype.UUID
		if err := rows.Scan(&p.ID, &p.UserID, &p.AmountCents, &p.Currency, &p.Status, &p.Type, &refID, &p.Description, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if refID.Valid {
			s := uuid.UUID(refID.Bytes).String()
			p.ReferenceID = &s
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *paymentRepo) CountPaidInMonth(ctx context.Context, year, month int) (int, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM payments WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2
	`, start, end).Scan(&n)
	return n, err
}

func (r *paymentRepo) TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	var total int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents), 0) FROM payments WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2
	`, start, end).Scan(&total)
	return total, err
}
