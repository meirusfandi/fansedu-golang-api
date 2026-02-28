package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) UserRepo {
	return &userRepo{pool: pool}
}

func (r *userRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	id := uuid.New().String()
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, name, role)
		VALUES ($1::uuid, $2, $3, $4, $5::user_role)
		RETURNING id, email, password_hash, name, role, avatar_url, email_verified_at, created_at, updated_at
	`, id, u.Email, u.PasswordHash, u.Name, u.Role)
	var out domain.User
	var avatarURL *string
	var emailVerifiedAt *time.Time
	err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.Name, &out.Role,
		&avatarURL, &emailVerifiedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.User{}, err
	}
	out.AvatarURL = avatarURL
	out.EmailVerifiedAt = emailVerifiedAt
	return out, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, avatar_url, email_verified_at, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	var out domain.User
	var avatarURL *string
	var emailVerifiedAt *time.Time
	err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.Name, &out.Role,
		&avatarURL, &emailVerifiedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.User{}, err
	}
	out.AvatarURL = avatarURL
	out.EmailVerifiedAt = emailVerifiedAt
	return out, nil
}

func (r *userRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, name, role, avatar_url, email_verified_at, created_at, updated_at
		FROM users WHERE id = $1::uuid
	`, id)
	var out domain.User
	var avatarURL *string
	var emailVerifiedAt *time.Time
	err := row.Scan(&out.ID, &out.Email, &out.PasswordHash, &out.Name, &out.Role,
		&avatarURL, &emailVerifiedAt, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return domain.User{}, err
	}
	out.AvatarURL = avatarURL
	out.EmailVerifiedAt = emailVerifiedAt
	return out, nil
}

func (r *userRepo) CountByRole(ctx context.Context, role string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role = $1::user_role`, role).Scan(&n)
	return n, err
}
