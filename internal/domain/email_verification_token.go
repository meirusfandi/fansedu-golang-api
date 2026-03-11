package domain

import "time"

type EmailVerificationToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

