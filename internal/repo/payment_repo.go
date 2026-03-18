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
	GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error)
	List(ctx context.Context, limit int) ([]domain.Payment, error)
	ListByUserID(ctx context.Context, userID string, limit int) ([]domain.Payment, error)
	GetByID(ctx context.Context, id string) (domain.Payment, error)
	Update(ctx context.Context, p domain.Payment) error
	CountPaidInMonth(ctx context.Context, year, month int) (int, error)
	TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
}

type paymentRepo struct{ pool *pgxpool.Pool }

func NewPaymentRepo(pool *pgxpool.Pool) PaymentRepo {
	return &paymentRepo{pool: pool}
}

func (r *paymentRepo) Create(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	id := uuid.New().String()
	var refArg, proofURL, confirmedBy, orderID interface{}
	if p.ReferenceID != nil {
		if u, err := uuid.Parse(*p.ReferenceID); err == nil {
			refArg = u
		}
	}
	if p.ProofURL != nil {
		proofURL = *p.ProofURL
	}
	if p.ConfirmedBy != nil {
		if u, err := uuid.Parse(*p.ConfirmedBy); err == nil {
			confirmedBy = u
		}
	}
	if p.OrderID != nil {
		if u, err := uuid.Parse(*p.OrderID); err == nil {
			orderID = u
		}
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (id, user_id, order_id, amount, currency, status, type, gateway, transaction_id, reference_id, description, proof_url, paid_at, confirmed_by, confirmed_at, rejection_note)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6::payment_status, $7::payment_type, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`, id, p.UserID, orderID, p.Amount, p.Currency, p.Status, p.Type, p.Gateway, p.TransactionID, refArg, p.Description, proofURL, p.PaidAt, confirmedBy, p.ConfirmedAt, p.RejectionNote).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Payment{}, err
	}
	p.ID = id
	return p, nil
}

func (r *paymentRepo) GetByOrderID(ctx context.Context, orderID string) (domain.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, order_id, amount, currency, status, type, gateway, transaction_id, reference_id, description, proof_url, paid_at, confirmed_by, confirmed_at, rejection_note, created_at, updated_at
		FROM payments WHERE order_id = $1::uuid ORDER BY created_at DESC LIMIT 1
	`, orderID)
	var p domain.Payment
	var refID, confirmedBy, ordID pgtype.UUID
	var proofURL, rejectionNote pgtype.Text
	var gateway, transactionID pgtype.Text
	err := row.Scan(&p.ID, &p.UserID, &ordID, &p.Amount, &p.Currency, &p.Status, &p.Type, &gateway, &transactionID, &refID, &p.Description, &proofURL, &p.PaidAt, &confirmedBy, &p.ConfirmedAt, &rejectionNote, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Payment{}, err
	}
	if refID.Valid {
		s := uuid.UUID(refID.Bytes).String()
		p.ReferenceID = &s
	}
	if ordID.Valid {
		s := uuid.UUID(ordID.Bytes).String()
		p.OrderID = &s
	}
	if proofURL.Valid {
		p.ProofURL = &proofURL.String
	}
	if confirmedBy.Valid {
		s := uuid.UUID(confirmedBy.Bytes).String()
		p.ConfirmedBy = &s
	}
	if rejectionNote.Valid {
		p.RejectionNote = &rejectionNote.String
	}
	if gateway.Valid {
		p.Gateway = &gateway.String
	}
	if transactionID.Valid {
		p.TransactionID = &transactionID.String
	}
	return p, nil
}

func (r *paymentRepo) List(ctx context.Context, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, currency, status, type, reference_id, description, proof_url, paid_at, confirmed_by, confirmed_at, rejection_note, created_at, updated_at
		FROM payments ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var refID, confirmedBy pgtype.UUID
		var proofURL, rejectionNote pgtype.Text
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Type, &refID, &p.Description, &proofURL, &p.PaidAt, &confirmedBy, &p.ConfirmedAt, &rejectionNote, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if refID.Valid {
			s := uuid.UUID(refID.Bytes).String()
			p.ReferenceID = &s
		}
		if proofURL.Valid {
			p.ProofURL = &proofURL.String
		}
		if confirmedBy.Valid {
			s := uuid.UUID(confirmedBy.Bytes).String()
			p.ConfirmedBy = &s
		}
		if rejectionNote.Valid {
			p.RejectionNote = &rejectionNote.String
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *paymentRepo) ListByUserID(ctx context.Context, userID string, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, currency, status, type, reference_id, description, proof_url, paid_at, confirmed_by, confirmed_at, rejection_note, created_at, updated_at
		FROM payments WHERE user_id = $1::uuid ORDER BY created_at DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var refID, confirmedBy pgtype.UUID
		var proofURL, rejectionNote pgtype.Text
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Type, &refID, &p.Description, &proofURL, &p.PaidAt, &confirmedBy, &p.ConfirmedAt, &rejectionNote, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if refID.Valid {
			s := uuid.UUID(refID.Bytes).String()
			p.ReferenceID = &s
		}
		if proofURL.Valid {
			p.ProofURL = &proofURL.String
		}
		if confirmedBy.Valid {
			s := uuid.UUID(confirmedBy.Bytes).String()
			p.ConfirmedBy = &s
		}
		if rejectionNote.Valid {
			p.RejectionNote = &rejectionNote.String
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *paymentRepo) GetByID(ctx context.Context, id string) (domain.Payment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, amount, currency, status, type, reference_id, description, proof_url, paid_at, confirmed_by, confirmed_at, rejection_note, created_at, updated_at
		FROM payments WHERE id = $1::uuid
	`, id)
	var p domain.Payment
	var refID, confirmedBy pgtype.UUID
	var proofURL, rejectionNote pgtype.Text
	err := row.Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Type, &refID, &p.Description, &proofURL, &p.PaidAt, &confirmedBy, &p.ConfirmedAt, &rejectionNote, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.Payment{}, err
	}
	if refID.Valid {
		s := uuid.UUID(refID.Bytes).String()
		p.ReferenceID = &s
	}
	if proofURL.Valid {
		p.ProofURL = &proofURL.String
	}
	if confirmedBy.Valid {
		s := uuid.UUID(confirmedBy.Bytes).String()
		p.ConfirmedBy = &s
	}
	if rejectionNote.Valid {
		p.RejectionNote = &rejectionNote.String
	}
	return p, nil
}

func (r *paymentRepo) Update(ctx context.Context, p domain.Payment) error {
	var proofURL, confirmedBy interface{}
	if p.ProofURL != nil {
		proofURL = *p.ProofURL
	}
	if p.ConfirmedBy != nil {
		if u, err := uuid.Parse(*p.ConfirmedBy); err == nil {
			confirmedBy = u
		}
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET status = $2::payment_status, paid_at = $3, proof_url = $4, confirmed_by = $5::uuid, confirmed_at = $6, rejection_note = $7, updated_at = NOW()
		WHERE id = $1::uuid
	`, p.ID, p.Status, p.PaidAt, proofURL, confirmedBy, p.ConfirmedAt, p.RejectionNote)
	return err
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
		SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2
	`, start, end).Scan(&total)
	return total, err
}
