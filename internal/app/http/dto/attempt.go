package dto

import "time"

type AttemptResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"userId"`
	TryoutSessionID  string     `json:"tryoutSessionId"`
	StartedAt        time.Time  `json:"startedAt"`
	SubmittedAt      *time.Time `json:"submittedAt,omitempty"`
	Status           string     `json:"status"`
	Score            *float64   `json:"score,omitempty"`
	MaxScore         *float64   `json:"maxScore,omitempty"`
	Percentile       *float64   `json:"percentile,omitempty"`
	TimeSecondsSpent *int       `json:"timeSecondsSpent,omitempty"`
	// Hanya terisi jika status submitted — untuk refresh halaman / muat ulang detail attempt.
	Review         []AttemptReviewRow  `json:"review,omitempty"`
	ModuleAnalysis []ModuleAnalysisRow `json:"moduleAnalysis,omitempty"`
	ModuleSummary  []ModuleAnalysisRow `json:"moduleSummary,omitempty"` // alias isi sama dengan moduleAnalysis (kompatibilitas FE)
}

type AnswerPutRequest struct {
	AnswerText     *string `json:"answerText,omitempty"`
	SelectedOption *string `json:"selectedOption,omitempty"`
	IsMarked       *bool   `json:"isMarked,omitempty"`
}

type SubmitResponse struct {
	AttemptID       string                 `json:"attemptId"`
	Score           float64                `json:"score"`
	MaxScore        float64                `json:"maxScore"`
	Percentile      float64                `json:"percentile"`
	Feedback        *FeedbackResponse      `json:"feedback,omitempty"`
	Review          []AttemptReviewRow     `json:"review,omitempty"`
	ModuleAnalysis  []ModuleAnalysisRow    `json:"moduleAnalysis,omitempty"`
	ModuleSummary   []ModuleAnalysisRow    `json:"moduleSummary,omitempty"` // alias isi sama (kompatibilitas FE)
}

// AttemptReviewRow pembahasan per soal setelah submit (isCorrect null = belum dinilai otomatis).
type AttemptReviewRow struct {
	QuestionID   string   `json:"questionId"`
	IsCorrect    *bool    `json:"isCorrect"` // tanpa omitempty agar null eksplisit di JSON
	ScoreGot     float64  `json:"scoreGot"`
	MaxScore     float64  `json:"maxScore"`
	ModuleKey    string   `json:"moduleKey,omitempty"`
	ModuleLabel  string   `json:"moduleLabel,omitempty"`
	ModuleID     *string  `json:"moduleId,omitempty"`
	ModuleTitle  *string  `json:"moduleTitle,omitempty"`
	Bidang       *string  `json:"bidang,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// ModuleAnalysisRow agregat benar/salah per modul/topik.
type ModuleAnalysisRow struct {
	ModuleKey      string `json:"moduleKey"`
	ModuleLabel    string `json:"moduleLabel"`
	QuestionCount  int    `json:"questionCount"`
	CorrectCount   int    `json:"correctCount"`
	WrongCount     int    `json:"wrongCount"`
	UnscoredCount  int    `json:"unscoredCount"`
}

type FeedbackResponse struct {
	Summary            *string  `json:"summary,omitempty"`
	Recap              *string  `json:"recap,omitempty"`
	StrengthAreas      []string `json:"strengthAreas,omitempty"`
	ImprovementAreas   []string `json:"improvementAreas,omitempty"`
	RecommendationText *string  `json:"recommendationText,omitempty"`
}

type QuestionResponse struct {
	ID              string      `json:"id"`
	TryoutSessionID string      `json:"tryoutSessionId"`
	SortOrder       int         `json:"sortOrder"`
	Type            string      `json:"type"`
	Body            string      `json:"body"`
	ImageURL        *string     `json:"imageUrl,omitempty"`
	ImageURLs       []string    `json:"imageUrls,omitempty"`
	Options         interface{} `json:"options,omitempty"`
	MaxScore        float64     `json:"maxScore"`
	ModuleID        *string     `json:"moduleId,omitempty"`
	ModuleTitle     *string     `json:"moduleTitle,omitempty"`
	Bidang          *string     `json:"bidang,omitempty"`
	Tags            []string    `json:"tags,omitempty"`
	CorrectOption   *string     `json:"correctOption,omitempty"` // hanya admin / penyusun soal
	CorrectText     *string     `json:"correctText,omitempty"`   // hanya admin
}
