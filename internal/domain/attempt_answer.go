package domain

import "time"

type AttemptAnswer struct {
	ID              string
	AttemptID       string
	QuestionID      string
	AnswerText      *string
	SelectedOption  *string
	IsMarked        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
