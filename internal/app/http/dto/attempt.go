package dto

import "time"

type AttemptResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	TryoutSessionID  string     `json:"tryout_session_id"`
	StartedAt        time.Time  `json:"started_at"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	Status           string     `json:"status"`
	Score            *float64   `json:"score,omitempty"`
	MaxScore         *float64   `json:"max_score,omitempty"`
	Percentile       *float64   `json:"percentile,omitempty"`
	TimeSecondsSpent *int       `json:"time_seconds_spent,omitempty"`
}

type AnswerPutRequest struct {
	AnswerText     *string `json:"answer_text,omitempty"`
	SelectedOption *string `json:"selected_option,omitempty"`
	IsMarked       *bool   `json:"is_marked,omitempty"`
}

type SubmitResponse struct {
	AttemptID string         `json:"attempt_id"`
	Score     float64        `json:"score"`
	Percentile float64       `json:"percentile"`
	Feedback  *FeedbackResponse `json:"feedback,omitempty"`
}

type FeedbackResponse struct {
	Summary            *string  `json:"summary,omitempty"`
	Recap              *string  `json:"recap,omitempty"`
	StrengthAreas       []string `json:"strength_areas,omitempty"`
	ImprovementAreas    []string `json:"improvement_areas,omitempty"`
	RecommendationText  *string  `json:"recommendation_text,omitempty"`
}

type QuestionResponse struct {
	ID              string          `json:"id"`
	TryoutSessionID string          `json:"tryout_session_id"`
	SortOrder       int             `json:"sort_order"`
	Type            string          `json:"type"`
	Body            string          `json:"body"`
	Options         interface{}     `json:"options,omitempty"`
	MaxScore        float64         `json:"max_score"`
}
