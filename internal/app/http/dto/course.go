package dto

type CourseResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Price       int     `json:"price"`
	Thumbnail   *string `json:"thumbnail,omitempty"`
	SubjectID   *string `json:"subjectId,omitempty"`
	CreatedBy   *string `json:"createdBy,omitempty"`
}

type EnrollmentResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"userId"`
	CourseID   string `json:"courseId"`
	Status     string `json:"status"`
	EnrolledAt string `json:"enrolledAt"`
}
