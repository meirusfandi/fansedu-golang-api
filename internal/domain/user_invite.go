package domain

import "time"

type UserInvite struct {
	ID        string
	UserID    string
	OrderID   *string
	Email     string
	Name      string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
