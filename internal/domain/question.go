package domain

import (
	"encoding/json"
	"time"
)

const (
	QuestionTypeShort          = "short"
	QuestionTypeMultipleChoice = "multiple_choice"
	QuestionTypeTrueFalse       = "true_false"
)

type Question struct {
	ID               string
	TryoutSessionID  string
	SortOrder        int
	Type             string
	Body             string
	Options          json.RawMessage // JSONB: ["A","B","C","D"] or [{"key":"A","label":"..."}]
	MaxScore         float64
	CreatedAt        time.Time
}
