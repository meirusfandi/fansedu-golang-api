package dto

type CourseResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	PriceCents  int     `json:"price_cents"`
	Thumbnail   *string `json:"thumbnail,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
}

type EnrollmentResponse struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	CourseID   string  `json:"course_id"`
	Status     string  `json:"status"`
	EnrolledAt string  `json:"enrolled_at"`
}
