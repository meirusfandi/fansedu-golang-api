package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var ErrPromoNotFound = pgx.ErrNoRows

// ErrVoucherClaimDuplicate user sudah punya klaim untuk promo ini.
var ErrVoucherClaimDuplicate = errors.New("voucher already claimed")

// ErrPromoCodeDuplicate kode promo sudah dipakai.
var ErrPromoCodeDuplicate = errors.New("promo code already exists")

type PromoRepo interface {
	GetByCode(ctx context.Context, code string) (domain.PromoCode, error)
	GetByID(ctx context.Context, id string) (domain.PromoCode, error)
	List(ctx context.Context) ([]domain.PromoCode, error)
	Create(ctx context.Context, p domain.PromoCode) (domain.PromoCode, error)
	Update(ctx context.Context, p domain.PromoCode) error
	Delete(ctx context.Context, id string) error
	IncrementUsedCount(ctx context.Context, id string) error
	// Klaim voucher (requires_claim).
	InsertClaim(ctx context.Context, userID, promoCodeID string) error
	HasUnusedClaim(ctx context.Context, userID, promoCodeID string) (bool, error)
	MarkClaimUsedForOrder(ctx context.Context, userID, promoCodeID, orderID string) error
	// ListMyUnusedVouchers klaim belum dipakai + kode promo untuk UI.
	ListMyUnusedVouchers(ctx context.Context, userID string) ([]UserVoucherListRow, error)
}

// UserVoucherListRow satu baris voucher yang sudah diklaim user, belum dipakai checkout.
type UserVoucherListRow struct {
	ClaimID       string
	PromoID       string
	Code          string
	DiscountType  string
	DiscountValue int
	ValidUntil    *time.Time
}

type promoRepo struct{ pool *pgxpool.Pool }

func NewPromoRepo(pool *pgxpool.Pool) PromoRepo {
	return &promoRepo{pool: pool}
}

func scanPromo(row pgx.Row) (domain.PromoCode, error) {
	var p domain.PromoCode
	var validUntil *time.Time
	var maxUses *int
	err := row.Scan(
		&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.ValidFrom, &validUntil, &maxUses, &p.UsedCount,
		&p.IsActive, &p.RequiresClaim, &p.AppliesToCourses, &p.AppliesToPackages,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return domain.PromoCode{}, err
	}
	p.ValidUntil = validUntil
	p.MaxUses = maxUses
	return p, nil
}

func (r *promoRepo) GetByCode(ctx context.Context, code string) (domain.PromoCode, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, discount_type, discount_value, valid_from, valid_until, max_uses, used_count,
		       COALESCE(is_active, TRUE), COALESCE(requires_claim, FALSE),
		       COALESCE(applies_to_courses, TRUE), COALESCE(applies_to_packages, TRUE),
		       created_at, updated_at
		FROM promo_codes
		WHERE LOWER(TRIM(code)) = LOWER(TRIM($1))
	`, code)
	return scanPromo(row)
}

func (r *promoRepo) GetByID(ctx context.Context, id string) (domain.PromoCode, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, discount_type, discount_value, valid_from, valid_until, max_uses, used_count,
		       COALESCE(is_active, TRUE), COALESCE(requires_claim, FALSE),
		       COALESCE(applies_to_courses, TRUE), COALESCE(applies_to_packages, TRUE),
		       created_at, updated_at
		FROM promo_codes WHERE id = $1::uuid
	`, id)
	return scanPromo(row)
}

func (r *promoRepo) List(ctx context.Context) ([]domain.PromoCode, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, discount_type, discount_value, valid_from, valid_until, max_uses, used_count,
		       COALESCE(is_active, TRUE), COALESCE(requires_claim, FALSE),
		       COALESCE(applies_to_courses, TRUE), COALESCE(applies_to_packages, TRUE),
		       created_at, updated_at
		FROM promo_codes ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.PromoCode
	for rows.Next() {
		p, err := scanPromo(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *promoRepo) Create(ctx context.Context, p domain.PromoCode) (domain.PromoCode, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO promo_codes (
			id, code, discount_type, discount_value, valid_from, valid_until, max_uses,
			is_active, requires_claim, applies_to_courses, applies_to_packages
		) VALUES (
			$1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING used_count, created_at, updated_at
	`, id, p.Code, p.DiscountType, p.DiscountValue, p.ValidFrom, p.ValidUntil, p.MaxUses,
		p.IsActive, p.RequiresClaim, p.AppliesToCourses, p.AppliesToPackages,
	).Scan(&p.UsedCount, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return domain.PromoCode{}, ErrPromoCodeDuplicate
		}
		return domain.PromoCode{}, err
	}
	p.ID = id
	return p, nil
}

func (r *promoRepo) Update(ctx context.Context, p domain.PromoCode) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE promo_codes SET
			code = $2, discount_type = $3, discount_value = $4, valid_from = $5, valid_until = $6, max_uses = $7,
			is_active = $8, requires_claim = $9, applies_to_courses = $10, applies_to_packages = $11,
			updated_at = NOW()
		WHERE id = $1::uuid
	`, p.ID, p.Code, p.DiscountType, p.DiscountValue, p.ValidFrom, p.ValidUntil, p.MaxUses,
		p.IsActive, p.RequiresClaim, p.AppliesToCourses, p.AppliesToPackages)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return ErrPromoCodeDuplicate
		}
		return err
	}
	return nil
}

func (r *promoRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM promo_codes WHERE id = $1::uuid`, id)
	return err
}

func (r *promoRepo) IncrementUsedCount(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE promo_codes SET used_count = used_count + 1, updated_at = NOW() WHERE id = $1::uuid
	`, id)
	return err
}

func (r *promoRepo) InsertClaim(ctx context.Context, userID, promoCodeID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO voucher_claims (user_id, promo_code_id) VALUES ($1::uuid, $2::uuid)
	`, userID, promoCodeID)
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return ErrVoucherClaimDuplicate
		}
		return err
	}
	return nil
}

func (r *promoRepo) HasUnusedClaim(ctx context.Context, userID, promoCodeID string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT 1 FROM voucher_claims
		WHERE user_id = $1::uuid AND promo_code_id = $2::uuid AND used_at IS NULL
		LIMIT 1
	`, userID, promoCodeID).Scan(&n)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *promoRepo) MarkClaimUsedForOrder(ctx context.Context, userID, promoCodeID, orderID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE voucher_claims SET used_at = NOW(), order_id = $3::uuid
		WHERE user_id = $1::uuid AND promo_code_id = $2::uuid AND used_at IS NULL
	`, userID, promoCodeID, orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *promoRepo) ListMyUnusedVouchers(ctx context.Context, userID string) ([]UserVoucherListRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT vc.id::text, p.id::text, p.code, p.discount_type, p.discount_value, p.valid_until
		FROM voucher_claims vc
		INNER JOIN promo_codes p ON p.id = vc.promo_code_id
		WHERE vc.user_id = $1::uuid AND vc.used_at IS NULL
		ORDER BY vc.claimed_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []UserVoucherListRow
	for rows.Next() {
		var row UserVoucherListRow
		if err := rows.Scan(&row.ClaimID, &row.PromoID, &row.Code, &row.DiscountType, &row.DiscountValue, &row.ValidUntil); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
