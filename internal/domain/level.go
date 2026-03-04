package domain

import "time"

type Level struct {
	ID          string
	Name        string
	Slug        string
	Description *string
	SortOrder   int
	IconURL     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
