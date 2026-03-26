package dto

// RoleItem for public GET /roles
type RoleItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	UserRoleCode string `json:"userRoleCode"`
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
	ReadAt    string `json:"readAt,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// UserPaymentResponse for GET /payments (my payments)
type UserPaymentResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userId"`
	Amount      int     `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	ReferenceID string  `json:"referenceId,omitempty"`
	Description *string `json:"description,omitempty"`
	ProofURL    string  `json:"proofUrl,omitempty"`
	PaidAt      string  `json:"paidAt,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}

// CourseMessageItem for course chat
type CourseMessageItem struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

// DiscussionItem for course discussions
type DiscussionItem struct {
	ID        string `json:"id"`
	CourseID  string `json:"courseId"`
	UserID    string `json:"userId"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// DiscussionReplyItem for discussion replies
type DiscussionReplyItem struct {
	ID           string `json:"id"`
	DiscussionID string `json:"discussionId"`
	UserID       string `json:"userId"`
	Body         string `json:"body"`
	CreatedAt    string `json:"createdAt"`
}

// CourseItem for list responses
type CourseItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	SubjectID   *string `json:"subjectId,omitempty"`
	CreatedBy   *string `json:"createdBy,omitempty"`
	CreatedAt   string  `json:"createdAt"`
}
