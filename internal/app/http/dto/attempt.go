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
}

type AnswerPutRequest struct {
	AnswerText     *string `json:"answerText,omitempty"`
	SelectedOption *string `json:"selectedOption,omitempty"`
	IsMarked       *bool   `json:"isMarked,omitempty"`
}

type SubmitResponse struct {
	AttemptID  string            `json:"attemptId"`
	Score      float64           `json:"score"`
	Percentile float64           `json:"percentile"`
	Feedback   *FeedbackResponse `json:"feedback,omitempty"`
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
}
