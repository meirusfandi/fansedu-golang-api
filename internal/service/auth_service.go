package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role string) (domain.User, string, error)
	RegisterWithInvite(ctx context.Context, token, email, name, password string) (domain.User, string, error)
	Login(ctx context.Context, email, password string) (domain.User, string, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	SetPassword(ctx context.Context, userID, newPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error
}

