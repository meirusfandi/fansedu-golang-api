package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type UserInviteRepo interface {
	Create(ctx context.Context, inv domain.UserInvite) (domain.UserInvite, error)
	GetByToken(ctx context.Context, token string) (domain.UserInvite, error)
	MarkUsed(ctx context.Context, id string, usedAt interface{}) error
	GetByOrderID(ctx context.Context, orderID string) (domain.UserInvite, error)
}

type userInviteRepo struct{ pool *pgxpool.Pool }

func NewUserInviteRepo(pool *pgxpool.Pool) UserInviteRepo { return &userInviteRepo{pool: pool} }

func (r *userInviteRepo) Create(ctx context.Context, inv domain.UserInvite) (domain.UserInvite, error) {
	id := uuid.New().String()
	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_invites (id, user_id, order_id, email, name, token, expires_at)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7)
		RETURNING created_at
	`, id, inv.UserID, inv.OrderID, inv.Email, inv.Name, inv.Token, inv.ExpiresAt).Scan(&inv.CreatedAt)
	if err != nil {
		return domain.UserInvite{}, err
	}
	inv.ID = id
	return inv, nil
}

func (r *userInviteRepo) GetByToken(ctx context.Context, token string) (domain.UserInvite, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, order_id, email, name, token, expires_at, used_at, created_at
		FROM user_invites WHERE token = $1
	`, token)
	var inv domain.UserInvite
	var orderID *string
	var usedAt *time.Time
	err := row.Scan(&inv.ID, &inv.UserID, &orderID, &inv.Email, &inv.Name, &inv.Token, &inv.ExpiresAt, &usedAt, &inv.CreatedAt)
	if err != nil {
		return domain.UserInvite{}, err
	}
	inv.OrderID = orderID
	inv.UsedAt = usedAt
	return inv, nil
}

func (r *userInviteRepo) MarkUsed(ctx context.Context, id string, _ interface{}) error {
	_, err := r.pool.Exec(ctx, `UPDATE user_invites SET used_at = NOW() WHERE id = $1::uuid`, id)
	return err
}

func (r *userInviteRepo) GetByOrderID(ctx context.Context, orderID string) (domain.UserInvite, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, order_id, email, name, token, expires_at, used_at, created_at
		FROM user_invites WHERE order_id = $1::uuid ORDER BY created_at DESC LIMIT 1
	`, orderID)
	var inv domain.UserInvite
	var oid *string
	var usedAt *time.Time
	err := row.Scan(&inv.ID, &inv.UserID, &oid, &inv.Email, &inv.Name, &inv.Token, &inv.ExpiresAt, &usedAt, &inv.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.UserInvite{}, err
		}
		return domain.UserInvite{}, err
	}
	inv.OrderID = oid
	inv.UsedAt = usedAt
	return inv, nil
}
