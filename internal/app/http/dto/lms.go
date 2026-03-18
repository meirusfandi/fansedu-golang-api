package dto

// RoleItem for public GET /roles
type RoleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SchoolItem for public GET /schools
type SchoolItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// NotificationItem for GET /notifications
type NotificationItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	ReadAt    string `json:"read_at,omitempty"`
	CreatedAt string `json:"created_at"`
}

// UserPaymentResponse for GET /payments (my payments)
type UserPaymentResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Amount       int    `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	ReferenceID  string `json:"reference_id,omitempty"`
	Description  *string `json:"description,omitempty"`
	ProofURL     string `json:"proof_url,omitempty"`
	PaidAt       string `json:"paid_at,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// CourseMessageItem for course chat
type CourseMessageItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// DiscussionItem for course discussions
type DiscussionItem struct {
	ID        string `json:"id"`
	CourseID  string `json:"course_id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// DiscussionReplyItem for discussion replies
type DiscussionReplyItem struct {
	ID            string `json:"id"`
	DiscussionID  string `json:"discussion_id"`
	UserID        string `json:"user_id"`
	Body          string `json:"body"`
	CreatedAt     string `json:"created_at"`
}

// CourseItem for list responses
type CourseItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
	CreatedBy   *string `json:"created_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
}
