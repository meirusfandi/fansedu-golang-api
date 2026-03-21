package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GeneratePasswordHash generates bcrypt hash using default app cost.
func GeneratePasswordHash(password string) (string, error) {
	return GeneratePasswordHashWithCost(password, bcrypt.DefaultCost)
}

// GeneratePasswordHashWithCost generates bcrypt hash with custom cost.
func GeneratePasswordHashWithCost(password string, cost int) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", errors.New("password is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
