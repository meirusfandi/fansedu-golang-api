package domain

import "time"

type Certificate struct {
	ID               string
	UserID           string
	TryoutSessionID  *string
	CourseID         *string
	IssuedAt         time.Time
	CreatedAt        time.Time
}
