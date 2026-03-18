package repo

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var ErrPromoNotFound = pgx.ErrNoRows

type PromoRepo interface {
	GetByCode(ctx context.Context, code string) (domain.PromoCode, error)
	IncrementUsedCount(ctx context.Context, id string) error
}

type promoRepo struct{ pool *pgxpool.Pool }

func NewPromoRepo(pool *pgxpool.Pool) PromoRepo {
	return &promoRepo{pool: pool}
}

func (r *promoRepo) GetByCode(ctx context.Context, code string) (domain.PromoCode, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, discount_type, discount_value, valid_from, valid_until, max_uses, used_count, created_at, updated_at
		FROM promo_codes
		WHERE LOWER(TRIM(code)) = LOWER(TRIM($1))
	`, code)
	var p domain.PromoCode
	var validUntil *time.Time
	var maxUses *int
	err := row.Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.ValidFrom, &validUntil, &maxUses, &p.UsedCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.PromoCode{}, err
	}
	p.ValidUntil = validUntil
	p.MaxUses = maxUses
	return p, nil
}

func (r *promoRepo) IncrementUsedCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE promo_codes SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1::uuid
	`, id)
	return err
}
