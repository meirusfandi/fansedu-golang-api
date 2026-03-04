package domain

import "time"

type Role struct {
	ID          string
	Name        string
	Slug        string
	Description *string
	IconURL     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
