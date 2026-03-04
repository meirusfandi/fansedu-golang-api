package dto

import "time"

type TryoutResponse struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	ShortTitle       *string    `json:"short_title,omitempty"`
	Description      *string    `json:"description,omitempty"`
	DurationMinutes  int        `json:"duration_minutes"`
	QuestionsCount   int        `json:"questions_count"`
	Level            string     `json:"level"`
	SubjectID        *string    `json:"subject_id,omitempty"`
	OpensAt          time.Time  `json:"opens_at"`
	ClosesAt         time.Time  `json:"closes_at"`
	MaxParticipants  *int       `json:"max_participants,omitempty"`
	Status           string     `json:"status"`
}

type TryoutCreateRequest struct {
	Title            string    `json:"title"`
	ShortTitle       *string   `json:"short_title,omitempty"`
	Description      *string   `json:"description,omitempty"`
	DurationMinutes  int       `json:"duration_minutes"`
	QuestionsCount   int       `json:"questions_count"`
	Level            string    `json:"level"`
	SubjectID        *string   `json:"subject_id,omitempty"`
	OpensAt          time.Time `json:"opens_at"`
	ClosesAt         time.Time `json:"closes_at"`
	MaxParticipants  *int      `json:"max_participants,omitempty"`
	Status           string    `json:"status"`
}

type TryoutStartResponse struct {
	AttemptID       string    `json:"attempt_id"`
	ExpiresAt       time.Time `json:"expires_at"`
	TimeLeftSeconds int       `json:"time_left_seconds"`
}
