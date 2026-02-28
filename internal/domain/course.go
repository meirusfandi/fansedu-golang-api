package domain

import "time"

type Course struct {
	ID          string
	Title       string
	Description *string
	CreatedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
