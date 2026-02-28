package domain

import "time"

const (
	TryoutLevelEasy   = "easy"
	TryoutLevelMedium = "medium"
	TryoutLevelHard   = "hard"

	TryoutStatusDraft  = "draft"
	TryoutStatusOpen   = "open"
	TryoutStatusClosed = "closed"
)

type TryoutSession struct {
	ID               string
	Title            string
	ShortTitle       *string
	Description      *string
	DurationMinutes  int
	QuestionsCount   int
	Level            string
	OpensAt          time.Time
	ClosesAt         time.Time
	MaxParticipants  *int
	Status           string
	CreatedBy        *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
