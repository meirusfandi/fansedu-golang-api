package domain

import "time"

type School struct {
	ID          string
	Name        string
	Slug        string
	Description *string
	Address     *string
	LogoURL     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
