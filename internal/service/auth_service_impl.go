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
	ErrEmailExists       = errors.New("email already registered")
	ErrInvalidCreds      = errors.New("invalid email or password")
	ErrAlreadyVerified   = errors.New("already verified")
	ErrInviteInvalid     = errors.New("invite token tidak valid atau sudah kadaluarsa")
	ErrInviteAlreadyUsed = errors.New("invite token sudah digunakan")
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
	inviteRepo interface {
		GetByToken(ctx context.Context, token string) (domain.UserInvite, error)
		MarkUsed(ctx context.Context, id string, usedAt interface{}) error
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
}, inviteRepo interface {
	GetByToken(ctx context.Context, token string) (domain.UserInvite, error)
	MarkUsed(ctx context.Context, id string, usedAt interface{}) error
}, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:       userRepo,
		emailTokenRepo: emailTokenRepo,
		inviteRepo:     inviteRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      24 * time.Hour,
		bcryptCost:     bcrypt.DefaultCost,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password, role string) (domain.User, string, error) {
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		// Email sudah terdaftar: perlakukan sebagai update password (reset/first setup),
		// bukan error. Dipakai untuk flow guest checkout yang kemudian isi password.
		hash, err := GeneratePasswordHashWithCost(password, s.bcryptCost)
		if err != nil {
			return domain.User{}, "", err
		}
		existing.PasswordHash = hash
		if strings.TrimSpace(name) != "" {
			existing.Name = name
		}
		now := time.Now()
		existing.EmailVerified = true
		if existing.EmailVerifiedAt == nil {
			existing.EmailVerifiedAt = &now
		}
		existing.MustSetPassword = false
		if err := s.userRepo.Update(ctx, existing); err != nil {
			return domain.User{}, "", err
		}
		token, err := s.signJWT(existing.ID, existing.Role)
		if err != nil {
			return domain.User{}, "", err
		}
		return existing, token, nil
	}
	role = strings.TrimSpace(role)
	if role == "" {
		role = domain.UserRoleStudent
	}
	hash, err := GeneratePasswordHashWithCost(password, s.bcryptCost)
	if err != nil {
		return domain.User{}, "", err
	}
	now := time.Now()
	u := domain.User{
		Email:           email,
		PasswordHash:    hash,
		Name:            name,
		Role:            role,
		EmailVerified:   true,
		EmailVerifiedAt: &now,
	}
	u, err = s.userRepo.Create(ctx, u)
	if err != nil {
		return domain.User{}, "", err
	}

	token, err := s.signJWT(u.ID, u.Role)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, token, nil
}

func (s *authService) RegisterWithInvite(ctx context.Context, token, email, name, password string) (domain.User, string, error) {
	if s.inviteRepo == nil {
		return domain.User{}, "", ErrInviteInvalid
	}
	inv, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return domain.User{}, "", ErrInviteInvalid
	}
	if inv.UsedAt != nil {
		return domain.User{}, "", ErrInviteAlreadyUsed
	}
	if time.Now().After(inv.ExpiresAt) {
		return domain.User{}, "", ErrInviteInvalid
	}
	if strings.TrimSpace(strings.ToLower(email)) != strings.TrimSpace(strings.ToLower(inv.Email)) {
		return domain.User{}, "", ErrInviteInvalid
	}
	u, err := s.userRepo.FindByID(ctx, inv.UserID)
	if err != nil {
		return domain.User{}, "", ErrInviteInvalid
	}
	hash, err := GeneratePasswordHashWithCost(password, s.bcryptCost)
	if err != nil {
		return domain.User{}, "", err
	}
	u.PasswordHash = hash
	if name != "" {
		u.Name = name
	}
	u.EmailVerified = true
	now := time.Now()
	u.EmailVerifiedAt = &now
	if err := s.userRepo.Update(ctx, u); err != nil {
		return domain.User{}, "", err
	}
	_ = s.inviteRepo.MarkUsed(ctx, inv.ID, nil)
	jwtToken, err := s.signJWT(u.ID, u.Role)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, jwtToken, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (domain.User, string, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, "", ErrInvalidCreds
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, "", ErrInvalidCreds
	}

	token, err := s.signJWT(u.ID, u.Role)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, token, nil
}

func (s *authService) signJWT(userID, role string) (string, error) {
	role = strings.TrimSpace(role)
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
	hash, err := GeneratePasswordHashWithCost(newPassword, s.bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.userRepo.Update(ctx, u)
}

func (s *authService) SetPassword(ctx context.Context, userID, newPassword string) error {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrInvalidCreds
	}
	hash, err := GeneratePasswordHashWithCost(newPassword, s.bcryptCost)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	u.MustSetPassword = false
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

