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

// LeaderboardTopRow — response GET .../leaderboard/top (Redis sorted set + nama user).
type LeaderboardTopRow struct {
	Rank     int     `json:"rank"`
	UserID   string  `json:"user_id"`
	UserName string  `json:"user_name"`
	Score    float64 `json:"score"`
}

// LeaderboardRankResponse — GET .../leaderboard/rank (user terautentikasi).
type LeaderboardRankResponse struct {
	InLeaderboard bool    `json:"in_leaderboard"`
	Rank          int     `json:"rank,omitempty"` // 1-based; hanya jika in_leaderboard
	Score         float64 `json:"score,omitempty"`
}
