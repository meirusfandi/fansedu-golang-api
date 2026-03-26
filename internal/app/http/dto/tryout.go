package dto

import "time"

type TryoutResponse struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	ShortTitle       *string   `json:"shortTitle,omitempty"`
	Description      *string   `json:"description,omitempty"`
	DurationMinutes  int       `json:"durationMinutes"`
	QuestionsCount   int       `json:"questionsCount"`
	Level            string    `json:"level"`
	SubjectID        *string   `json:"subjectId,omitempty"`
	OpensAt          time.Time `json:"opensAt"`
	ClosesAt         time.Time `json:"closesAt"`
	MaxParticipants  *int      `json:"maxParticipants,omitempty"`
	Status           string    `json:"status"`
}

type TryoutCreateRequest struct {
	Title            string    `json:"title"`
	ShortTitle       *string   `json:"shortTitle,omitempty"`
	Description      *string   `json:"description,omitempty"`
	DurationMinutes  int       `json:"durationMinutes"`
	QuestionsCount   int       `json:"questionsCount"`
	Level            string    `json:"level"`
	SubjectID        *string   `json:"subjectId,omitempty"`
	OpensAt          time.Time `json:"opensAt"`
	ClosesAt         time.Time `json:"closesAt"`
	MaxParticipants  *int      `json:"maxParticipants,omitempty"`
	Status           string    `json:"status"`
}

type TryoutStartResponse struct {
	AttemptID       string    `json:"attemptId"`
	ExpiresAt       time.Time `json:"expiresAt"`
	TimeLeftSeconds int       `json:"timeLeftSeconds"`
}

// LeaderboardTopRow — response GET .../leaderboard/top
type LeaderboardTopRow struct {
	Rank     int     `json:"rank"`
	UserID   string  `json:"userId"`
	UserName string  `json:"userName"`
	Score    float64 `json:"score"`
}

// LeaderboardRankResponse — GET .../leaderboard/rank
type LeaderboardRankResponse struct {
	InLeaderboard bool    `json:"inLeaderboard"`
	Rank          int     `json:"rank,omitempty"`
	Score         float64 `json:"score,omitempty"`
}
