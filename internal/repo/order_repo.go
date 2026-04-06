package repo

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type OrderRepo interface {
	Create(ctx context.Context, o domain.Order) (domain.Order, error)
	GetByID(ctx context.Context, id string) (domain.Order, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Order, error)
	ListByUserIDWithFilters(ctx context.Context, userID, status, search string, page, limit int) ([]domain.Order, int, error)
	GetPendingByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Order, bool, error)
	GetPendingByUserAndPackage(ctx context.Context, userID, packageID string) (domain.Order, bool, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string, proofAt *time.Time) error
	UpdateOrderCreatedAt(ctx context.Context, orderID string, createdAt time.Time) error
	UpdatePaymentProofAtOnly(ctx context.Context, orderID string, proofAt time.Time) error
}

type orderRepo struct{ pool *pgxpool.Pool }

func NewOrderRepo(pool *pgxpool.Pool) OrderRepo { return &orderRepo{pool: pool} }

func (r *orderRepo) Create(ctx context.Context, o domain.Order) (domain.Order, error) {
	id := uuid.New().String()
	normalPrice := o.NormalPrice
	if normalPrice == 0 && o.TotalPrice > 0 {
		normalPrice = o.TotalPrice + o.Discount
	}
	quantity := o.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	subtotal := o.Subtotal
	if subtotal <= 0 {
		subtotal = normalPrice
	}
	unitPrice := o.UnitPrice
	if unitPrice <= 0 {
		unitPrice = subtotal / quantity
		if unitPrice <= 0 {
			unitPrice = normalPrice
		}
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO orders (
			id, user_id, status, total_price, normal_price, quantity, unit_price, subtotal, unique_code, is_collective, students_json,
			promo_code, discount, discount_percent, confirmation_code, payment_method, payment_reference, role_hint, buyer_email, package_id
		)
		VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13, $14, $15, $16, $17, $18, $19, $20::uuid)
		RETURNING created_at, updated_at
	`, id, o.UserID, o.Status, o.TotalPrice, normalPrice, quantity, unitPrice, subtotal, o.UniqueCode, o.IsCollective, o.StudentsJSON, o.PromoCode, o.Discount, o.DiscountPercent, o.ConfirmationCode, o.PaymentMethod, o.PaymentReference, o.RoleHint, o.BuyerEmail, o.PackageID).Scan(&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}
	o.ID = id
	o.NormalPrice = normalPrice
	o.Quantity = quantity
	o.UnitPrice = unitPrice
	o.Subtotal = subtotal
	return o, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id string) (domain.Order, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, status, total_price, COALESCE(normal_price, total_price), COALESCE(quantity, 1), COALESCE(unit_price, COALESCE(normal_price, total_price)),
		       COALESCE(subtotal, total_price), COALESCE(unique_code, 0), COALESCE(is_collective, false), COALESCE(students_json, '[]'::jsonb),
		       promo_code, COALESCE(discount, 0), discount_percent, confirmation_code, payment_method, payment_reference,
		       payment_proof_url, payment_proof_at, sender_account_no, sender_name, role_hint, buyer_email, package_id, created_at, updated_at
		FROM orders WHERE id = $1::uuid
	`, id)
	var o domain.Order
	var promoCode, confCode *string
	var discountPercent *float64
	var packageID *string
	err := row.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.NormalPrice, &o.Quantity, &o.UnitPrice, &o.Subtotal, &o.UniqueCode, &o.IsCollective, &o.StudentsJSON, &promoCode, &o.Discount, &discountPercent, &confCode, &o.PaymentMethod, &o.PaymentReference,
		&o.PaymentProofURL, &o.PaymentProofAt, &o.SenderAccountNo, &o.SenderName, &o.RoleHint, &o.BuyerEmail, &packageID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return domain.Order{}, err
	}
	o.PromoCode = promoCode
	o.DiscountPercent = discountPercent
	o.ConfirmationCode = confCode
	o.PackageID = packageID
	return o, nil
}

func (r *orderRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, status, total_price, COALESCE(normal_price, total_price), COALESCE(quantity, 1), COALESCE(unit_price, COALESCE(normal_price, total_price)),
		       COALESCE(subtotal, total_price), COALESCE(unique_code, 0), COALESCE(is_collective, false), COALESCE(students_json, '[]'::jsonb),
		       promo_code, COALESCE(discount, 0), discount_percent, confirmation_code, payment_method, payment_reference, package_id, created_at, updated_at
		FROM orders WHERE user_id = $1::uuid ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Order
	for rows.Next() {
		var o domain.Order
		var promoCode, confCode *string
		var discountPercent *float64
		var packageID *string
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.NormalPrice, &o.Quantity, &o.UnitPrice, &o.Subtotal, &o.UniqueCode, &o.IsCollective, &o.StudentsJSON, &promoCode, &o.Discount, &discountPercent, &confCode, &o.PaymentMethod, &o.PaymentReference, &packageID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.PromoCode = promoCode
		o.DiscountPercent = discountPercent
		o.ConfirmationCode = confCode
		o.PackageID = packageID
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *orderRepo) ListByUserIDWithFilters(ctx context.Context, userID, status, search string, page, limit int) ([]domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := `
		SELECT
			o.id,
			o.user_id,
			o.status,
			o.total_price,
			COALESCE(o.normal_price, o.total_price),
			COALESCE(o.quantity, 1),
			COALESCE(o.unit_price, COALESCE(o.normal_price, o.total_price)),
			COALESCE(o.subtotal, o.total_price),
			COALESCE(o.unique_code, 0),
			COALESCE(o.is_collective, false),
			COALESCE(o.students_json, '[]'::jsonb),
			o.promo_code,
			COALESCE(o.discount, 0),
			o.discount_percent,
			o.confirmation_code,
			o.package_id,
			o.created_at,
			o.updated_at,
			COUNT(*) OVER()::int AS total
		FROM orders o
		WHERE o.user_id = $1::uuid
		  AND (
			CASE
				WHEN $2 = '' THEN true
				ELSE o.status = $2::order_status
			END
		  )
		  AND (
			$3 = '' OR
			o.id::text ILIKE '%' || $3 || '%' OR
			EXISTS (
				SELECT 1
				FROM order_items oi
				JOIN courses c ON c.id = oi.course_id
				WHERE oi.order_id = o.id
				  AND (c.title ILIKE '%' || $3 || '%' OR COALESCE(c.slug, '') ILIKE '%' || $3 || '%')
			)
		  )
		ORDER BY o.created_at DESC
		LIMIT $4 OFFSET $5
	`

	rows, err := r.pool.Query(ctx, query, userID, status, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]domain.Order, 0, limit)
	total := 0
	for rows.Next() {
		var o domain.Order
		var promoCode, confCode *string
		var discountPercent *float64
		var packageID *string
		var rowTotal int
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.NormalPrice, &o.Quantity, &o.UnitPrice, &o.Subtotal, &o.UniqueCode, &o.IsCollective, &o.StudentsJSON,
			&promoCode, &o.Discount, &discountPercent, &confCode,
			&packageID,
			&o.CreatedAt, &o.UpdatedAt, &rowTotal,
		); err != nil {
			return nil, 0, err
		}
		o.PromoCode = promoCode
		o.DiscountPercent = discountPercent
		o.ConfirmationCode = confCode
		o.PackageID = packageID
		out = append(out, o)
		total = rowTotal
	}
	return out, total, rows.Err()
}

func (r *orderRepo) GetPendingByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Order, bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT o.id, o.user_id, o.status, o.total_price, COALESCE(o.normal_price, o.total_price), COALESCE(o.quantity, 1),
		       COALESCE(o.unit_price, COALESCE(o.normal_price, o.total_price)), COALESCE(o.subtotal, o.total_price),
		       COALESCE(o.unique_code, 0), COALESCE(o.is_collective, false), COALESCE(o.students_json, '[]'::jsonb),
		       o.promo_code, COALESCE(o.discount, 0), o.discount_percent, o.confirmation_code, o.payment_method, o.payment_reference, o.created_at, o.updated_at
		FROM orders o
		JOIN order_items oi ON oi.order_id = o.id AND oi.course_id = $2::uuid
		WHERE o.user_id = $1::uuid AND o.status = 'pending'
		ORDER BY o.created_at DESC
		LIMIT 1
	`, userID, courseID)
	var o domain.Order
	var promoCode, confCode *string
	var discountPercent *float64
	err := row.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.NormalPrice, &o.Quantity, &o.UnitPrice, &o.Subtotal, &o.UniqueCode, &o.IsCollective, &o.StudentsJSON, &promoCode, &o.Discount, &discountPercent, &confCode, &o.PaymentMethod, &o.PaymentReference, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, false, nil
		}
		return domain.Order{}, false, err
	}
	o.PromoCode = promoCode
	o.DiscountPercent = discountPercent
	o.ConfirmationCode = confCode
	return o, true, nil
}

func (r *orderRepo) GetPendingByUserAndPackage(ctx context.Context, userID, packageID string) (domain.Order, bool, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT o.id, o.user_id, o.status, o.total_price, COALESCE(o.normal_price, o.total_price), COALESCE(o.quantity, 1),
		       COALESCE(o.unit_price, COALESCE(o.normal_price, o.total_price)), COALESCE(o.subtotal, o.total_price),
		       COALESCE(o.unique_code, 0), COALESCE(o.is_collective, false), COALESCE(o.students_json, '[]'::jsonb),
		       o.promo_code, COALESCE(o.discount, 0), o.discount_percent, o.confirmation_code, o.payment_method, o.payment_reference,
		       o.package_id, o.created_at, o.updated_at
		FROM orders o
		WHERE o.user_id = $1::uuid AND o.status = 'pending' AND o.package_id = $2::uuid
		ORDER BY o.created_at DESC
		LIMIT 1
	`, userID, packageID)
	var o domain.Order
	var promoCode, confCode *string
	var discountPercent *float64
	var pkgID *string
	err := row.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.NormalPrice, &o.Quantity, &o.UnitPrice, &o.Subtotal, &o.UniqueCode, &o.IsCollective, &o.StudentsJSON, &promoCode, &o.Discount, &discountPercent, &confCode, &o.PaymentMethod, &o.PaymentReference,
		&pkgID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, false, nil
		}
		return domain.Order{}, false, err
	}
	o.PromoCode = promoCode
	o.DiscountPercent = discountPercent
	o.ConfirmationCode = confCode
	o.PackageID = pkgID
	return o, true, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1::uuid`, id, status)
	return err
}

func (r *orderRepo) UpdatePaymentProof(ctx context.Context, orderID, proofURL, senderAccountNo, senderName string, proofAt *time.Time) error {
	var at interface{}
	if proofAt != nil {
		at = *proofAt
	} else {
		at = nil // COALESCE below → NOW()
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE orders SET payment_proof_url = $2, payment_proof_at = COALESCE($5::timestamptz, NOW()),
			sender_account_no = $3, sender_name = $4, updated_at = NOW()
		WHERE id = $1::uuid
	`, orderID, nullIfEmpty(proofURL), nullIfEmpty(senderAccountNo), nullIfEmpty(senderName), at)
	return err
}

func (r *orderRepo) UpdateOrderCreatedAt(ctx context.Context, orderID string, createdAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE orders SET created_at = $2, updated_at = NOW() WHERE id = $1::uuid
	`, orderID, createdAt)
	return err
}

func (r *orderRepo) UpdatePaymentProofAtOnly(ctx context.Context, orderID string, proofAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE orders SET payment_proof_at = $2, updated_at = NOW() WHERE id = $1::uuid
	`, orderID, proofAt)
	return err
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
