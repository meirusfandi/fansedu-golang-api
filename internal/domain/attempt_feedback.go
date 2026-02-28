package domain

import (
	"encoding/json"
	"time"
)

type AttemptFeedback struct {
	ID                 string
	AttemptID           string
	Summary            *string
	Recap              *string
	StrengthAreas       json.RawMessage // JSONB array of strings
	ImprovementAreas   json.RawMessage
	RecommendationText *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
