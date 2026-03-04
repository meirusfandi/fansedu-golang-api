package domain

import "time"

type Event struct {
	ID           string
	Title        string
	Slug         string
	Description  *string
	StartAt      time.Time
	EndAt        time.Time
	ThumbnailURL *string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
