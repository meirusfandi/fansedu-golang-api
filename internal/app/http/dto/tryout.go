package dto

import "time"

// GuruTryoutPaperResponse — GET /api/v1/guru/tryouts/:id/paper (dan trainer setara).
type GuruTryoutPaperResponse struct {
	Title            string              `json:"title"`
	DurationMinutes  int                 `json:"durationMinutes"`
	OpensAt          time.Time           `json:"opensAt"`
	ClosesAt         time.Time           `json:"closesAt"`
	Questions        []QuestionResponse  `json:"questions"`
}

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
	// GradingMode: "auto" (kunci + otomatis) | "manual" (skor dari admin/guru lewat review).
	GradingMode      string    `json:"gradingMode"`
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
	GradingMode      string    `json:"gradingMode,omitempty"`
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
