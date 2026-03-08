package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AuthService interface {
	Register(ctx context.Context, name, email, password, role string) (domain.User, string, error)
	Login(ctx context.Context, email, password string) (domain.User, string, error)
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
}

