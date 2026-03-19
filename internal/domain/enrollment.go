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

// StudentCourseRow is a denormalized row for listing student's courses with progress.
type StudentCourseRow struct {
	EnrollmentID     string
	CourseID         string
	CourseTitle      string
	CourseSlug       string
	CourseThumbnail  string
	EnrollmentStatus string
	EnrolledAt       time.Time
}
