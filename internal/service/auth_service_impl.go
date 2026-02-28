package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var (
	ErrEmailExists   = errors.New("email already registered")
	ErrInvalidCreds  = errors.New("invalid email or password")
)

type authService struct {
	userRepo   interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		FindByEmail(ctx context.Context, email string) (domain.User, error)
	}
	jwtSecret  []byte
	jwtExpiry  time.Duration
	bcryptCost int
}

func NewAuthService(userRepo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtSecret:  jwtSecret,
		jwtExpiry:  24 * time.Hour,
		bcryptCost: bcrypt.DefaultCost,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (domain.User, string, error) {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return domain.User{}, "", ErrEmailExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return domain.User{}, "", err
	}
	u := domain.User{
		Email:       email,
		PasswordHash: string(hash),
		Name:        name,
		Role:        domain.UserRoleStudent,
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
	claims := jwt.MapClaims{
		"sub": userID,
		"role": role,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.jwtExpiry).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}
