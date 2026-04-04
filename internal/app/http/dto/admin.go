package dto

type AdminOverviewResponse struct {
	TotalStudents     int     `json:"totalStudents"`
	TotalUsers        int     `json:"totalUsers"`
	ActiveTryouts     int     `json:"activeTryouts"`
	TotalCourses      int     `json:"totalCourses"`
	TotalEnrollments  int     `json:"totalEnrollments"`
	AvgScore          float64 `json:"avgScore"`
	TotalCertificates int     `json:"totalCertificates"`
}

type UserListResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	SchoolID  *string `json:"schoolId,omitempty"`
	SubjectID *string `json:"subjectId,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

type UserDetailResponse struct {
	ID        string           `json:"id"`
	Email     string           `json:"email"`
	Name      string           `json:"name"`
	Role      string           `json:"role"`
	AvatarURL *string          `json:"avatarUrl,omitempty"`
	SchoolID  *string          `json:"schoolId,omitempty"`
	SubjectID *string          `json:"subjectId,omitempty"`
	School    *SchoolResponse  `json:"school,omitempty"`
	Subject   *SubjectResponse `json:"subject,omitempty"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
}

type UserCreateRequest struct {
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	Name      string  `json:"name"`
	Role      string  `json:"role"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	SchoolID  *string `json:"schoolId,omitempty"`
	SubjectID *string `json:"subjectId,omitempty"`
}

type UserUpdateRequest struct {
	Email     *string `json:"email,omitempty"`
	Name      *string `json:"name,omitempty"`
	Role      *string `json:"role,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	SchoolID  *string `json:"schoolId,omitempty"`
	SubjectID *string `json:"subjectId,omitempty"`
	Password  *string `json:"password,omitempty"`
}

type QuestionCreateRequest struct {
	TryoutSessionID string      `json:"tryoutSessionId"`
	SortOrder       int         `json:"sortOrder"`
	Type            string      `json:"type"`
	Body            string      `json:"body"`
	ImageURL        *string     `json:"imageUrl,omitempty"`
	ImageURLs       []string    `json:"imageUrls,omitempty"`
	Options         interface{} `json:"options,omitempty"`
	MaxScore        float64     `json:"maxScore"`
	ModuleID        *string     `json:"moduleId,omitempty"`
	ModuleTitle     *string     `json:"moduleTitle,omitempty"`
	Bidang          *string     `json:"bidang,omitempty"`
	Tags            []string    `json:"tags,omitempty"`
	CorrectOption   *string     `json:"correctOption,omitempty"`
	CorrectText     *string     `json:"correctText,omitempty"`
}

type QuestionUpdateRequest struct {
	SortOrder     *int           `json:"sortOrder,omitempty"`
	Type          *string        `json:"type,omitempty"`
	Body          *string        `json:"body,omitempty"`
	ImageURL      *string        `json:"imageUrl,omitempty"`
	ImageURLs     *[]string      `json:"imageUrls,omitempty"`
	Options       *interface{}   `json:"options,omitempty"`
	MaxScore      *float64       `json:"maxScore,omitempty"`
	ModuleID      *string        `json:"moduleId,omitempty"`
	ModuleTitle   *string        `json:"moduleTitle,omitempty"`
	Bidang        *string        `json:"bidang,omitempty"`
	Tags          *[]string      `json:"tags,omitempty"`
	CorrectOption *string        `json:"correctOption,omitempty"`
	CorrectText   *string        `json:"correctText,omitempty"`
}

type CourseCreateRequest struct {
	Title       string  `json:"title"`
	Slug        *string `json:"slug,omitempty"`
	Description *string `json:"description,omitempty"`
	Price       *int    `json:"price,omitempty"`
	Thumbnail   *string `json:"thumbnail,omitempty"`
	SubjectID   *string `json:"subjectId,omitempty"`
}

type CertificateIssueRequest struct {
	UserID          string  `json:"userId"`
	TryoutSessionID *string `json:"tryoutSessionId,omitempty"`
	CourseID        *string `json:"courseId,omitempty"`
}

type CourseContentRequest struct {
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	SortOrder   int         `json:"sortOrder"`
	Type        string      `json:"type"`
	Content     interface{} `json:"content,omitempty"`
}

type CourseContentResponse struct {
	ID          string      `json:"id"`
	CourseID    string      `json:"courseId"`
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	SortOrder   int         `json:"sortOrder"`
	Type        string      `json:"type"`
	Content     interface{} `json:"content,omitempty"`
	CreatedAt   string      `json:"createdAt"`
}

// AdminCourseLinkedPackage — paket landing yang berisi kelas ini (package_courses).
type AdminCourseLinkedPackage struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// AdminCourseLinkedTryout — tryout terhubung ke kelas (course_tryouts).
type AdminCourseLinkedTryout struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	OpensAt   string `json:"opensAt"`
	ClosesAt  string `json:"closesAt"`
	SortOrder int    `json:"sortOrder"`
}

// AdminCourseManageResponse — ringkasan admin: kelas + konten + tautan paket & tryout.
type AdminCourseManageResponse struct {
	Course           CourseResponse              `json:"course"`
	Contents         []CourseContentResponse     `json:"contents"`
	ContentsByType   map[string]int              `json:"contentsByType"`
	LinkedPackages   []AdminCourseLinkedPackage  `json:"linkedPackages"`
	LinkedTryouts    []AdminCourseLinkedTryout   `json:"linkedTryouts"`
	RelatedEndpoints RelatedCourseAdminEndpoints `json:"relatedEndpoints"`
}

// RelatedCourseAdminEndpoints dokumentasi URL relatif base /api/v1/admin untuk FE.
type RelatedCourseAdminEndpoints struct {
	ListContents    string `json:"listContents"`
	CreateContent   string `json:"createContent"`
	UpdateContent   string `json:"updateContent"`
	DeleteContent   string `json:"deleteContent"`
	ListEnrollments string `json:"listEnrollments"`
	TryoutQuestions string `json:"tryoutQuestions"`
	PackageManage   string `json:"packageManageNote"`
	GetProgram      string `json:"getProgram"`
	PutProgram      string `json:"putProgram"`
}

// AdminCourseProgramMeetingItem satu pertemuan (1–8): judul, detail, PDF, PR, link live.
type AdminCourseProgramMeetingItem struct {
	MeetingNumber  int     `json:"meetingNumber"`
	Title          string  `json:"title"`
	DetailText     *string `json:"detailText,omitempty"`
	PdfURL         *string `json:"pdfUrl,omitempty"`
	PrTitle        *string `json:"prTitle,omitempty"`
	PrDescription  *string `json:"prDescription,omitempty"`
	LiveClassURL   *string `json:"liveClassUrl,omitempty"`
}

// AdminCourseProgramResponse GET .../courses/{courseId}/program
type AdminCourseProgramResponse struct {
	TrackType               string                          `json:"trackType"`
	Meetings                []AdminCourseProgramMeetingItem `json:"meetings"`
	PretestTryoutSessionID  *string                         `json:"pretestTryoutSessionId,omitempty"`
}

// AdminCourseProgramPutRequest PUT .../courses/{courseId}/program — setelah simpan, learning journey di-rebuild dari data ini.
type AdminCourseProgramPutRequest struct {
	TrackType               string                          `json:"trackType"`
	Meetings                []AdminCourseProgramMeetingItem `json:"meetings"`
	PretestTryoutSessionID  *string                         `json:"pretestTryoutSessionId,omitempty"`
}

// AdminCourseLinkedPackagesPutRequest body PUT .../linked-packages
type AdminCourseLinkedPackagesPutRequest struct {
	PackageIDs []string `json:"packageIds"`
}

// AdminCourseLinkedTryoutsPutRequest body PUT .../linked-tryouts
type AdminCourseLinkedTryoutsPutRequest struct {
	TryoutIDs []string `json:"tryoutIds"`
}

type PaymentCreateRequest struct {
	UserID      string  `json:"userId"`
	Amount      int     `json:"amount"`
	Currency    string  `json:"currency"`
	Type        string  `json:"type"`
	ReferenceID *string `json:"referenceId,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

type PaymentResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	Amount    int     `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
	Type      string  `json:"type"`
	PaidAt    *string `json:"paidAt,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

type MonthlyReportResponse struct {
	Year           int   `json:"year"`
	Month          int   `json:"month"`
	NewEnrollments int   `json:"newEnrollments"`
	PaymentsCount  int   `json:"paymentsCount"`
	TotalRevenue   int64 `json:"totalRevenue"`
}

// --- Role ---
type RoleRequest struct {
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	UserRoleCode *string `json:"userRoleCode,omitempty"`
	Description  *string `json:"description,omitempty"`
	IconURL      *string `json:"iconUrl,omitempty"`
}

type RoleResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	UserRoleCode string  `json:"userRoleCode"`
	Description  *string `json:"description,omitempty"`
	IconURL      *string `json:"iconUrl,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// --- School ---
type SchoolRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Address     *string `json:"address,omitempty"`
	LogoURL     *string `json:"logoUrl,omitempty"`
}

type SchoolResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Address     *string `json:"address,omitempty"`
	LogoURL     *string `json:"logoUrl,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// --- Setting ---
type SettingRequest struct {
	Key         string      `json:"key"`
	Slug        string      `json:"slug"`
	Value       *string     `json:"value,omitempty"`
	ValueJSON   interface{} `json:"valueJson,omitempty"`
	Description *string     `json:"description,omitempty"`
}

type SettingResponse struct {
	ID          string      `json:"id"`
	Key         string      `json:"key"`
	Slug        string      `json:"slug"`
	Value       *string     `json:"value,omitempty"`
	ValueJSON   interface{} `json:"valueJson,omitempty"`
	Description *string     `json:"description,omitempty"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
}

// --- Event ---
type EventRequest struct {
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	StartAt      string  `json:"startAt"`
	EndAt        string  `json:"endAt"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
	Status       string  `json:"status"`
}

type EventResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	StartAt      string  `json:"startAt"`
	EndAt        string  `json:"endAt"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

// --- Subject ---
type SubjectRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"iconUrl,omitempty"`
	SortOrder   int     `json:"sortOrder"`
}

type SubjectResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	IconURL     *string `json:"iconUrl,omitempty"`
	SortOrder   int     `json:"sortOrder"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// --- Level (jenjang pendidikan: SD, SMP, SMA) ---
type LevelRequest struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sortOrder"`
	IconURL     *string `json:"iconUrl,omitempty"`
}

type LevelResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	SortOrder   int     `json:"sortOrder"`
	IconURL     *string `json:"iconUrl,omitempty"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type LevelWithSubjectsResponse struct {
	LevelResponse
	Subjects []SubjectResponse `json:"subjects,omitempty"`
}

// --- Course Report (laporan rekap skor tryout, kehadiran, progress per kelas) ---
type CourseReportResponse struct {
	Course      CourseReportCourseInfo   `json:"course"`
	GeneratedAt string                   `json:"generatedAt"`
	Students    []CourseReportStudentRow `json:"students"`
}

type CourseReportCourseInfo struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type CourseReportStudentRow struct {
	StudentID         string                   `json:"studentId"`
	StudentName       string                   `json:"studentName"`
	StudentEmail      string                   `json:"studentEmail"`
	EnrolledAt        string                   `json:"enrolledAt"`
	EnrollmentStatus  string                   `json:"enrollmentStatus"`
	Progress          CourseReportProgress     `json:"progress"`
	TryoutScores      []CourseReportTryoutScore `json:"tryoutScores"`
	Attendance        CourseReportAttendance   `json:"attendance"`
}

type CourseReportProgress struct {
	Status      string  `json:"status"`
	CompletedAt *string `json:"completedAt,omitempty"`
}

type CourseReportTryoutScore struct {
	TryoutID    string   `json:"tryoutId"`
	TryoutTitle string   `json:"tryoutTitle"`
	AttemptID   string   `json:"attemptId"`
	Score       *float64 `json:"score,omitempty"`
	MaxScore    *float64 `json:"maxScore,omitempty"`
	Percentile  *float64 `json:"percentile,omitempty"`
	SubmittedAt *string  `json:"submittedAt,omitempty"`
}

type CourseReportAttendance struct {
	TryoutsParticipated int     `json:"tryoutsParticipated"`
	LastActivityAt      *string `json:"lastActivityAt,omitempty"`
}

// ErrorLogPatchRequest — PATCH /admin/error-logs/:id (errors.manage)
type ErrorLogPatchRequest struct {
	Resolved  *bool   `json:"resolved"`
	AdminNote *string `json:"adminNote,omitempty"`
}
