package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CertificateRepo interface {
	Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.Certificate, error)
	Count(ctx context.Context) (int, error)
}

type certificateRepo struct{ pool *pgxpool.Pool }

func NewCertificateRepo(pool *pgxpool.Pool) CertificateRepo { return &certificateRepo{pool: pool} }

func (r *certificateRepo) Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO certificates (id, user_id, tryout_session_id, course_id)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid)
		RETURNING issued_at, created_at
	`, id, c.UserID, c.TryoutSessionID, c.CourseID).Scan(&c.IssuedAt, &c.CreatedAt)
	if err != nil {
		return domain.Certificate{}, err
	}
	c.ID = id
	return c, nil
}

func (r *certificateRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, tryout_session_id, course_id, issued_at, created_at
		FROM certificates WHERE user_id = $1::uuid ORDER BY issued_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		if err := rows.Scan(&c.ID, &c.UserID, &c.TryoutSessionID, &c.CourseID, &c.IssuedAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *certificateRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM certificates`).Scan(&n)
	return n, err
}
