package domain

import (
	"encoding/json"
	"time"
)

type Setting struct {
	ID          string
	Key         string
	Slug        string
	Value       *string
	ValueJSON   json.RawMessage
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
