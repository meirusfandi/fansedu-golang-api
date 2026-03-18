package domain

import "time"

type Course struct {
	ID          string
	Title       string
	Slug        *string
	Description *string
	Price int // nominal dalam rupiah
	Thumbnail   *string
	SubjectID   *string
	CreatedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
