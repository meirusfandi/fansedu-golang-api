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
	UserID          string   `json:"userId"`
	UserName        string   `json:"userName"`
	SchoolName      *string  `json:"schoolName,omitempty"`
	BestScore       *float64 `json:"bestScore,omitempty"`
	BestTimeSeconds *int     `json:"bestTimeSeconds,omitempty"`
	HasAttempt      bool     `json:"hasAttempt"`
}
