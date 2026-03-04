package domain

import (
	"encoding/json"
	"time"
)

const (
	CourseContentTypeModule = "module"
	CourseContentTypeQuiz   = "quiz"
	CourseContentTypeTest   = "test"
)

type CourseContent struct {
	ID          string
	CourseID    string
	Title       string
	Description *string
	SortOrder   int
	Type        string
	Content     json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
