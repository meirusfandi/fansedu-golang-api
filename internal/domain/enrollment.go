package domain

import "time"

const (
	EnrollmentStatusEnrolled   = "enrolled"
	EnrollmentStatusInProgress = "in_progress"
	EnrollmentStatusCompleted  = "completed"
)

type CourseEnrollment struct {
	ID          string
	UserID      string
	CourseID    string
	Status      string
	EnrolledAt  time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
}
