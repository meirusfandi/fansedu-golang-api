package dto

type GenerateQuestionsRequest struct {
	Subject    string `json:"subject"`
	Grade      string `json:"grade"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	Count      int    `json:"count"`
}

type SubmitAnswerRequest struct {
	QuestionID  string `json:"questionId"`
	Answer      string `json:"answer"`
	TimeSpentMs int64  `json:"timeSpentMs"`
}

type CreateSubscriptionRequest struct {
	PlanCode string  `json:"planCode"`
	StartAt  *string `json:"startAt,omitempty"` // RFC3339
	EndAt    *string `json:"endAt,omitempty"`   // RFC3339
}

