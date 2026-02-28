package domain

import "time"

const (
	AttemptStatusInProgress = "in_progress"
	AttemptStatusSubmitted  = "submitted"
	AttemptStatusExpired    = "expired"
)

type Attempt struct {
	ID               string
	UserID           string
	TryoutSessionID  string
	StartedAt        time.Time
	SubmittedAt      *time.Time
	Status           string
	Score            *float64
	MaxScore         *float64
	Percentile       *float64
	TimeSecondsSpent *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
