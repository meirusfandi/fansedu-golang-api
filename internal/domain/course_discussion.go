package domain

import "time"

type CourseDiscussion struct {
	ID        string
	CourseID  string
	UserID    string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CourseDiscussionReply struct {
	ID            string
	DiscussionID  string
	UserID        string
	Body          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
