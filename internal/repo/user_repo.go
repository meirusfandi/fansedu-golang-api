package repo

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type UserRepo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	// FindByIDProfile: sama seperti FindByID tanpa membaca password_hash (lebih ringan untuk /auth/me & profil).
	FindByIDProfile(ctx context.Context, id string) (domain.User, error)
	// FindByIDProfileWithSchool: satu round-trip user + sekolah (trainer/instructor profile).
	FindByIDProfileWithSchool(ctx context.Context, id string) (domain.User, *domain.School, error)
	// MustSetPasswordByID: query minimal untuk PasswordSetupGuard.
	MustSetPasswordByID(ctx context.Context, id string) (bool, error)
	CountByRole(ctx context.Context, role string) (int, error)
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, role string) ([]domain.User, error)
	Update(ctx context.Context, u domain.User) error
}

