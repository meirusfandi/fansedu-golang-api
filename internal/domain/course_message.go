package domain

import "time"

type CourseMessage struct {
	ID        string
	CourseID  string
	UserID    string
	Message   string
	CreatedAt time.Time
}
