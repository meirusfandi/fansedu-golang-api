package domain

import "time"

const (
	UserRoleAdmin   = "admin"
	UserRoleStudent = "student"
	UserRoleGuru    = "guru"
	UserRoleTrainer = "trainer" // nanti dibuat oleh admin
)

type User struct {
	ID              string
	Email           string
	PasswordHash    string
	Name            string
	Role            string
	AvatarURL       *string
	SchoolID        *string
	SubjectID       *string
	EmailVerified   bool
	EmailVerifiedAt *time.Time
	MustSetPassword bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
