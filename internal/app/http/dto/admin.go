package dto

type AdminOverviewResponse struct {
	TotalStudents     int     `json:"total_students"`
	TotalUsers        int     `json:"total_users"`
	ActiveTryouts     int     `json:"active_tryouts"`
	TotalCourses      int     `json:"total_courses"`
	TotalEnrollments  int     `json:"total_enrollments"`
	AvgScore          float64 `json:"avg_score"`
	TotalCertificates int     `json:"total_certificates"`
}

type UserListResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	SchoolID  *string `json:"school_id,omitempty"`
	SubjectID *string `json:"subject_id,omitempty"`
	CreatedAt string  `json:"created_at"`
}

type UserDetailResponse struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	Name      string           `json:"name"`
	Role      string           `json:"role"`
	AvatarURL *string          `json:"avatar_url,omitempty"`
	SchoolID  *string          `json:"school_id,omitempty"`
	SubjectID *string          `json:"subject_id,omitempty"`
	School    *SchoolResponse  `json:"school,omitempty"`
	Subject   *SubjectResponse `json:"subject,omitempty"`
	CreatedAt string           `json:"created_at"`
	UpdatedAt string           `json:"updated_at"`
}

type UserCreateRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	SchoolID  *string `json:"school_id,omitempty"`
	SubjectID *string `json:"subject_id,omitempty"`
}

type UserUpdateRequest struct {
	Email     *string `json:"email,omitempty"`
	Name      *string `json:"name,omitempty"`
	Role      *string `json:"role,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	SchoolID  *string `json:"school_id,omitempty"`
	SubjectID *string `json:"subject_id,omitempty"`
	Password  *string `json:"password,omitempty"`
}

type QuestionCreateRequest struct {
	TryoutSessionID string   `json:"tryout_session_id"`
	SortOrder       int      `json:"sort_order"`
	Type            string   `json:"type"`
	Body            string   `json:"body"`             // Teks atau HTML isi soal (boleh berisi HTML + <img>)
	ImageURL        *string  `json:"image_url,omitempty"`
	ImageURLs       []string `json:"image_urls,omitempty"` // Array URL gambar tambahan
	Options         interface{} `json:"options,omitempty"`
	MaxScore        float64  `json:"max_score"`
}

type QuestionUpdateRequest struct {
	SortOrder  *int      `json:"sort_order,omitempty"`
	Type       *string   `json:"type,omitempty"`
	Body       *string   `json:"body,omitempty"`        // Teks atau HTML isi soal
	ImageURL   *string   `json:"image_url,omitempty"`
	ImageURLs  *[]string `json:"image_urls,omitempty"` // Array URL gambar (kirim null/omit untuk tidak ubah)
	Options    *interface{} `json:"options,omitempty"`
	MaxScore   *float64  `json:"max_score,omitempty"`
}

type CourseCreateRequest struct {
	Title       string  `json:"title"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	PriceCents  *int    `json:"price_cents,omitempty"`
	Thumbnail   *string `json:"thumbnail,omitempty"`
	SubjectID   *string `json:"subject_id,omitempty"`
}

type CertificateIssueRequest struct {
	UserID           string  `json:"user_id"`
	TryoutSessionID  *string `json:"tryout_session_id,omitempty"`
	CourseID         *string `json:"course_id,omitempty"`
}

type CourseContentRequest struct {
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	SortOrder   int         `json:"sort_order"`
	Type        string      `json:"type"`
	Content     interface{} `json:"content,omitempty"`
}

type CourseContentResponse struct {
	ID          string      `json:"id"`
	CourseID    string      `json:"course_id"`
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	SortOrder   int         `json:"sort_order"`
	Type        string      `json:"type"`
	Content     interface{} `json:"content,omitempty"`
	CreatedAt   string      `json:"created_at"`
}

type PaymentCreateRequest struct {
	UserID      string  `json:"user_id"`
	AmountCents int     `json:"amount_cents"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
	ReferenceID *string `json:"reference_id,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

type PaymentResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	AmountCents int     `json:"amount_cents"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	PaidAt      *string `json:"paid_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type MonthlyReportResponse struct {
	Year               int   `json:"year"`
	Month              int   `json:"month"`
	NewEnrollments     int   `json:"new_enrollments"`
	PaymentsCount      int   `json:"payments_count"`
	TotalRevenueCents  int64 `json:"total_revenue_cents"`
}

// --- Role ---
type RoleRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
}
type RoleResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// --- School ---
type SchoolRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Address     *string `json:"address,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
}
type SchoolResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Address     *string `json:"address,omitempty"`
	LogoURL     *string `json:"logo_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// --- Setting ---
type SettingRequest struct {
	Key         string      `json:"key"`
	Slug        string      `json:"slug"`
	Value       *string     `json:"value,omitempty"`
	ValueJSON   interface{} `json:"value_json,omitempty"`
	Description *string     `json:"description,omitempty"`
}
type SettingResponse struct {
	ID          string      `json:"id"`
	Key         string      `json:"key"`
	Slug        string      `json:"slug"`
	Value       *string     `json:"value,omitempty"`
	ValueJSON   interface{} `json:"value_json,omitempty"`
	Description *string     `json:"description,omitempty"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// --- Event ---
type EventRequest struct {
	Title        string   `json:"title"`
	Slug         string   `json:"slug"`
	Description  *string  `json:"description,omitempty"`
	StartAt      string   `json:"start_at"`
	EndAt        string   `json:"end_at"`
	ThumbnailURL *string  `json:"thumbnail_url,omitempty"`
	Status       string   `json:"status"`
}
type EventResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	StartAt      string  `json:"start_at"`
	EndAt        string  `json:"end_at"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// --- Subject ---
type SubjectRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	SortOrder   int     `json:"sort_order"`
}
type SubjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"icon_url,omitempty"`
	SortOrder   int     `json:"sort_order"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// --- Level (jenjang pendidikan: SD, SMP, SMA) ---
type LevelRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sort_order"`
	IconURL     *string `json:"icon_url,omitempty"`
}
type LevelResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sort_order"`
	IconURL     *string `json:"icon_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
type LevelWithSubjectsResponse struct {
	LevelResponse
	Subjects []SubjectResponse `json:"subjects,omitempty"`
}

// --- Course Report (laporan rekap skor tryout, kehadiran, progress per kelas) ---
type CourseReportResponse struct {
	Course      CourseReportCourseInfo   `json:"course"`
	GeneratedAt string                   `json:"generated_at"`
	Students    []CourseReportStudentRow `json:"students"`
}

type CourseReportCourseInfo struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type CourseReportStudentRow struct {
	StudentID        string                      `json:"student_id"`
	StudentName      string                      `json:"student_name"`
	StudentEmail     string                      `json:"student_email"`
	EnrolledAt       string                      `json:"enrolled_at"`
	EnrollmentStatus string                      `json:"enrollment_status"`
	Progress         CourseReportProgress        `json:"progress"`
	TryoutScores     []CourseReportTryoutScore    `json:"tryout_scores"`
	Attendance       CourseReportAttendance      `json:"attendance"`
}

type CourseReportProgress struct {
	Status      string  `json:"status"`
	CompletedAt *string `json:"completed_at,omitempty"`
}

type CourseReportTryoutScore struct {
	TryoutID    string   `json:"tryout_id"`
	TryoutTitle string   `json:"tryout_title"`
	AttemptID   string   `json:"attempt_id"`
	Score       *float64 `json:"score,omitempty"`
	MaxScore    *float64 `json:"max_score,omitempty"`
	Percentile  *float64 `json:"percentile,omitempty"`
	SubmittedAt *string `json:"submitted_at,omitempty"`
}

type CourseReportAttendance struct {
	TryoutsParticipated int     `json:"tryouts_participated"`
	LastActivityAt       *string `json:"last_activity_at,omitempty"`
}
