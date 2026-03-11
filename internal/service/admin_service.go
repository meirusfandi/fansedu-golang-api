package service

import (
	"context"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AdminOverview struct {
	TotalStudents     int     `json:"total_students"`
	TotalUsers        int     `json:"total_users"`
	ActiveTryouts     int     `json:"active_tryouts"`
	TotalCourses      int     `json:"total_courses"`
	TotalEnrollments  int     `json:"total_enrollments"`
	AvgScore          float64 `json:"avg_score"`
	TotalCertificates int     `json:"total_certificates"`
}

type AdminService interface {
	Overview(ctx context.Context) (*AdminOverview, error)
	ListUsers(ctx context.Context, role string) ([]domain.User, error)
	GetUser(ctx context.Context, id string) (domain.User, error)
	CreateUser(ctx context.Context, u domain.User, password string) (domain.User, error)
	UpdateUser(ctx context.Context, u domain.User, newPassword *string) error
	ListCourses(ctx context.Context) ([]domain.Course, error)
	GetCourseByID(ctx context.Context, id string) (domain.Course, error)
	UpdateCourse(ctx context.Context, c domain.Course) error
	ListTryouts(ctx context.Context) ([]domain.TryoutSession, error)
	CreateTryout(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	UpdateTryout(ctx context.Context, t domain.TryoutSession) error
	DeleteTryout(ctx context.Context, id string) error
	ListQuestions(ctx context.Context, tryoutID string) ([]domain.Question, error)
	GetQuestion(ctx context.Context, questionID string) (domain.Question, error)
	CreateQuestion(ctx context.Context, q domain.Question) (domain.Question, error)
	UpdateQuestion(ctx context.Context, q domain.Question) error
	DeleteQuestion(ctx context.Context, id string) error
	CreateCourse(ctx context.Context, c domain.Course) (domain.Course, error)
	ListEnrollmentsByCourse(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	IssueCertificate(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	ListCourseContents(ctx context.Context, courseID string) ([]domain.CourseContent, error)
	GetCourseContent(ctx context.Context, id string) (domain.CourseContent, error)
	CreateCourseContent(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
	UpdateCourseContent(ctx context.Context, c domain.CourseContent) error
	DeleteCourseContent(ctx context.Context, id string) error
	ListPayments(ctx context.Context, limit int) ([]domain.Payment, error)
	CreatePayment(ctx context.Context, p domain.Payment) (domain.Payment, error)
	ConfirmPayment(ctx context.Context, paymentID string, confirmed bool, adminID string, rejectionNote *string) error
	ReportMonthly(ctx context.Context, year, month int) (*MonthlyReport, error)
	GetCourseReport(ctx context.Context, courseID string) (*CourseReport, error)
	GetQuestionStats(ctx context.Context, tryoutID, questionID string) (*QuestionStats, error)
	GetTryoutQuestionStatsBulk(ctx context.Context, tryoutID string) (*QuestionStatsBulk, error)
	GetTryoutAnalysis(ctx context.Context, tryoutID string) (*TryoutAnalysis, error)
	ListTryoutStudents(ctx context.Context, tryoutID string) ([]TryoutStudentItem, error)
	GetAttemptAIAnalysis(ctx context.Context, tryoutID, attemptID string) (*AttemptAIAnalysisResponse, error)
}

type MonthlyReport struct {
	Year               int   `json:"year"`
	Month              int   `json:"month"`
	NewEnrollments     int   `json:"new_enrollments"`
	PaymentsCount      int   `json:"payments_count"`
	TotalRevenueCents  int64 `json:"total_revenue_cents"`
}

// QuestionStats response for GET /admin/tryouts/:tryoutId/questions/:questionId/stats
type QuestionStats struct {
	ParticipantsCount int     `json:"participants_count"`
	AnsweredCount     int     `json:"answered_count"`
	CorrectCount      int     `json:"correct_count"`
	WrongCount        int     `json:"wrong_count"`
	CorrectPercent    float64 `json:"correct_percent"`
	WrongPercent      float64 `json:"wrong_percent"`
}

// QuestionStatsItem one question in bulk stats
type QuestionStatsItem struct {
	QuestionID    string  `json:"question_id"`
	AnsweredCount int     `json:"answered_count"`
	CorrectCount  int     `json:"correct_count"`
	WrongCount    int     `json:"wrong_count"`
	CorrectPercent float64 `json:"correct_percent"`
	WrongPercent   float64 `json:"wrong_percent"`
}

// QuestionStatsBulk response for GET /admin/tryouts/:tryoutId/questions/stats
type QuestionStatsBulk struct {
	ParticipantsCount int                `json:"participants_count"`
	Questions         []QuestionStatsItem `json:"questions"`
}

// TryoutAnalysis response for GET /admin/tryouts/:tryoutId/analysis (grafik & analisis per soal)
type TryoutAnalysis struct {
	TryoutID           string                    `json:"tryout_id"`
	TryoutTitle        string                    `json:"tryout_title"`
	ParticipantsCount  int                       `json:"participants_count"`
	Questions          []TryoutAnalysisQuestion  `json:"questions"`
}

// TryoutAnalysisQuestion satu soal dalam analisis tryout (siap untuk grafik)
type TryoutAnalysisQuestion struct {
	QuestionNumber     int                `json:"question_number"`
	QuestionID         string             `json:"question_id"`
	QuestionType       string             `json:"question_type"`
	AnsweredCount     int                `json:"answered_count"`
	UnansweredCount   int                `json:"unanswered_count"`
	CorrectCount      int                `json:"correct_count"`
	WrongCount        int                `json:"wrong_count"`
	CorrectPercent    float64            `json:"correct_percent"`
	WrongPercent      float64            `json:"wrong_percent"`
	OptionDistribution map[string]int    `json:"option_distribution"` // A, B, C, D, ...
}

// TryoutStudentItem satu siswa yang submit tryout (untuk daftar + link ke analisis AI)
type TryoutStudentItem struct {
	UserID       string   `json:"user_id"`
	UserName     string   `json:"user_name"`
	UserEmail    string   `json:"user_email"`
	AttemptID    string   `json:"attempt_id"`
	Score        *float64 `json:"score"`
	MaxScore     *float64 `json:"max_score"`
	Percentile   *float64 `json:"percentile"`
	SubmittedAt  *string  `json:"submitted_at"`
}

// AttemptAIAnalysisResponse analisis AI per attempt (per siswa)
type AttemptAIAnalysisResponse struct {
	AttemptID        string   `json:"attempt_id"`
	Summary          string   `json:"summary"`
	Recap            string   `json:"recap"`
	StrengthAreas    []string `json:"strength_areas"`
	ImprovementAreas []string `json:"improvement_areas"`
	Recommendation   string   `json:"recommendation"`
}

// CourseReport laporan rekap skor tryout, kehadiran, progress siswa per kelas (course)
type CourseReport struct {
	Course      CourseReportCourse
	GeneratedAt time.Time
	Students    []CourseReportStudent
}

type CourseReportCourse struct {
	ID          string
	Title       string
	Description *string
}

type CourseReportStudent struct {
	StudentID        string
	StudentName      string
	StudentEmail     string
	EnrolledAt       time.Time
	EnrollmentStatus string
	Progress         CourseReportStudentProgress
	TryoutScores     []CourseReportTryoutScore
	Attendance       CourseReportStudentAttendance
}

type CourseReportStudentProgress struct {
	Status      string
	CompletedAt *time.Time
}

type CourseReportTryoutScore struct {
	TryoutID    string
	TryoutTitle string
	AttemptID   string
	Score       *float64
	MaxScore    *float64
	Percentile  *float64
	SubmittedAt *time.Time
}

type CourseReportStudentAttendance struct {
	TryoutsParticipated int
	LastActivityAt      *time.Time
}
