package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var (
	ErrEmailExists      = errors.New("email already registered")
	ErrInvalidCreds     = errors.New("invalid email or password")
	ErrEmailNotVerified = errors.New("email not verified")
	ErrAlreadyVerified  = errors.New("already verified")
)

type authService struct {
	userRepo interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		FindByEmail(ctx context.Context, email string) (domain.User, error)
		FindByID(ctx context.Context, id string) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
	}
	emailTokenRepo interface {
		Create(ctx context.Context, t domain.EmailVerificationToken) (domain.EmailVerificationToken, error)
		GetByToken(ctx context.Context, token string) (domain.EmailVerificationToken, error)
		MarkUsed(ctx context.Context, id string, usedAt time.Time) error
	}
	jwtSecret  []byte
	jwtExpiry  time.Duration
	bcryptCost int
}

func NewAuthService(userRepo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, u domain.User) error
}, emailTokenRepo interface {
	Create(ctx context.Context, t domain.EmailVerificationToken) (domain.EmailVerificationToken, error)
	GetByToken(ctx context.Context, token string) (domain.EmailVerificationToken, error)
	MarkUsed(ctx context.Context, id string, usedAt time.Time) error
}, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:       userRepo,
		emailTokenRepo: emailTokenRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      24 * time.Hour,
		bcryptCost:     bcrypt.DefaultCost,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password, role string) (domain.User, string, error) {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return domain.User{}, "", ErrEmailExists
	}
	role = normalizeRegisterRole(role)
	if role != domain.UserRoleStudent && role != domain.UserRoleGuru && role != "instructor" {
		role = domain.UserRoleStudent
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return domain.User{}, "", err
	}
	u := domain.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		EmailVerified: false,
	}
	u, err = s.userRepo.Create(ctx, u)
	if err != nil {
		return domain.User{}, "", err
	}

	// For student/guru/instructor, require email verification before login.
	if role == domain.UserRoleStudent || role == domain.UserRoleGuru || role == "instructor" {
		if s.emailTokenRepo != nil {
			_, _ = s.emailTokenRepo.Create(ctx, domain.EmailVerificationToken{
				ID:        uuid.New().String(),
				UserID:    u.ID,
				Token:     uuid.New().String(),
				ExpiresAt: time.Now().Add(24 * time.Hour),
			})
		}
		// Do not issue JWT yet; user must verify email first.
		return u, "", nil
	}

	token, err := s.signJWT(u.ID, u.Role)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, token, nil
}

func normalizeRegisterRole(r string) string {
	r = strings.TrimSpace(strings.ToLower(r))
	if r == "" || r == "siswa" {
		return domain.UserRoleStudent
	}
	if r == domain.UserRoleGuru || r == "instructor" {
		if r == "instructor" {
			return "instructor"
		}
		return domain.UserRoleGuru
	}
	return r
}

func (s *authService) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, "", ErrInvalidCreds
	}

	// Require email verification for student/guru/instructor
	if (u.Role == domain.UserRoleStudent || u.Role == domain.UserRoleGuru || u.Role == "instructor") && !u.EmailVerified {
		return domain.User{}, "", ErrEmailNotVerified
	}

	token, err := s.signJWT(u.ID, u.Role)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, token, nil
}

func (s *authService) signJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"role": role,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.jwtExpiry).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}

func (s *authService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCreds
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return s.userRepo.Update(ctx, u)
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	if s.emailTokenRepo == nil {
		return errors.New("email verification not configured")
	}
	t, err := s.emailTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return ErrInvalidCreds
	}
	if t.UsedAt != nil || time.Now().After(t.ExpiresAt) {
		return ErrInvalidCreds
	}
	u, err := s.userRepo.FindByID(ctx, t.UserID)
	if err != nil {
		return ErrInvalidCreds
	}
	now := time.Now()
	if !u.EmailVerified {
		u.EmailVerified = true
		if u.EmailVerifiedAt == nil {
			u.EmailVerifiedAt = &now
		}
		if err := s.userRepo.Update(ctx, u); err != nil {
			return err
		}
	}
	if err := s.emailTokenRepo.MarkUsed(ctx, t.ID, now); err != nil {
		return err
	}
	return nil
}

func (s *authService) ResendVerification(ctx context.Context, email string) error {
	if s.emailTokenRepo == nil {
		return errors.New("email verification not configured")
	}
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Jangan bocorkan apakah email ada atau tidak
		return nil
	}
	if u.EmailVerified {
		return ErrAlreadyVerified
	}
	_, err = s.emailTokenRepo.Create(ctx, domain.EmailVerificationToken{
		ID:        uuid.New().String(),
		UserID:    u.ID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	return err
}

