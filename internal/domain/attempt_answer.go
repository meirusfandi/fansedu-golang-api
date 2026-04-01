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
	// Review manual (admin / trainer): jika ManualScore terisi, menggantikan skor otomatis untuk soal ini.
	ManualScore       *float64
	ReviewerComment   *string
	ReviewedByUserID  *string
	ReviewedAt        *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
