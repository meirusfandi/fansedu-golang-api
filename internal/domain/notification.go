package domain

import "time"

type Notification struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	Type      string
	ReadAt    *time.Time
	CreatedAt time.Time
}
