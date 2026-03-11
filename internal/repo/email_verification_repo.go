package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type EmailVerificationTokenRepo interface {
	Create(ctx context.Context, t domain.EmailVerificationToken) (domain.EmailVerificationToken, error)
	GetByToken(ctx context.Context, token string) (domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
}

type emailVerificationTokenRepo struct {
	pool *pgxpool.Pool
}

func NewEmailVerificationTokenRepo(pool *pgxpool.Pool) EmailVerificationTokenRepo {
	return &emailVerificationTokenRepo{pool: pool}
}

func (r *emailVerificationTokenRepo) Create(ctx context.Context, t domain.EmailVerificationToken) (domain.EmailVerificationToken, error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at)
		VALUES ($1::uuid, $2::uuid, $3, $4)
	`, t.ID, t.UserID, t.Token, t.ExpiresAt)
	if err != nil {
		return domain.EmailVerificationToken{}, err
	}
	return t, nil
}

func (r *emailVerificationTokenRepo) GetByToken(ctx context.Context, token string) (domain.EmailVerificationToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM email_verification_tokens
		WHERE token = $1
	`, token)
	var t domain.EmailVerificationToken
	err := row.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	return t, err
}

func (r *emailVerificationTokenRepo) MarkUsed(ctx context.Context, id string, usedAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE email_verification_tokens
		SET used_at = $2
		WHERE id = $1::uuid
	`, id, usedAt)
	return err
}

