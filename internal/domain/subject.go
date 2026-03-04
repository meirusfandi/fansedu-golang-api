package domain

import "time"

type Subject struct {
	ID          string
	Name        string
	Slug        string
	Description *string
	IconURL     *string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
