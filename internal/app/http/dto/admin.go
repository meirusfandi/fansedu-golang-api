package dto

type AdminOverviewResponse struct {
	TotalStudents     int     `json:"total_students"`
	ActiveTryouts     int     `json:"active_tryouts"`
	AvgScore          float64 `json:"avg_score"`
	TotalCertificates int     `json:"total_certificates"`
}

type QuestionCreateRequest struct {
	TryoutSessionID string      `json:"tryout_session_id"`
	SortOrder       int         `json:"sort_order"`
	Type            string      `json:"type"`
	Body            string      `json:"body"`
	Options         interface{} `json:"options,omitempty"`
	MaxScore        float64     `json:"max_score"`
}

type QuestionUpdateRequest struct {
	SortOrder *int     `json:"sort_order,omitempty"`
	Type      *string  `json:"type,omitempty"`
	Body      *string  `json:"body,omitempty"`
	Options   *interface{} `json:"options,omitempty"`
	MaxScore  *float64 `json:"max_score,omitempty"`
}

type CourseCreateRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type CertificateIssueRequest struct {
	UserID           string  `json:"user_id"`
	TryoutSessionID  *string `json:"tryout_session_id,omitempty"`
	CourseID         *string `json:"course_id,omitempty"`
}
