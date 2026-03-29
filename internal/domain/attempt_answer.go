package domain

import "time"

type AttemptAnswer struct {
	ID              string
	AttemptID       string
	QuestionID      string
	AnswerText      *string
	SelectedOption  *string
	IsMarked        bool
	IsCorrect       *bool // diisi setelah submit; nil = belum dinilai otomatis
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
