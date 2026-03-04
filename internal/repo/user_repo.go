package repo

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type UserRepo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	CountByRole(ctx context.Context, role string) (int, error)
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, role string) ([]domain.User, error)
	Update(ctx context.Context, u domain.User) error
}

