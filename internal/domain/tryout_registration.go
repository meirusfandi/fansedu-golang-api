package domain

import "time"

type TryoutRegistration struct {
	ID               string
	UserID           string
	TryoutSessionID  string
	RegisteredAt     time.Time
}

// LeaderboardEntry untuk satu baris di leaderboard tryout (nama siswa, sekolah, nilai)
type LeaderboardEntry struct {
	Rank            int      `json:"rank"`
	UserID          string   `json:"user_id"`
	UserName        string   `json:"user_name"`
	SchoolName      *string  `json:"school_name,omitempty"`
	BestScore       *float64 `json:"best_score,omitempty"`
	BestTimeSeconds *int     `json:"best_time_seconds,omitempty"`
	HasAttempt      bool     `json:"has_attempt"`
}
