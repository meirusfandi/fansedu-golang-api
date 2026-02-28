package domain

import "time"

const (
	UserRoleAdmin   = "admin"
	UserRoleStudent = "student"
)

type User struct {
	ID              string
	Email           string
	PasswordHash     string
	Name            string
	Role            string
	AvatarURL       *string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
