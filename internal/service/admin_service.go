package service

import (
	"context"
	"errors"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AdminOverview struct {
	TotalStudents     int     `json:"totalStudents"`
	TotalUsers        int     `json:"totalUsers"`
	ActiveTryouts     int     `json:"activeTryouts"`
	TotalCourses      int     `json:"totalCourses"`
	TotalEnrollments  int     `json:"totalEnrollments"`
	AvgScore          float64 `json:"avgScore"`
	TotalCertificates int     `json:"totalCertificates"`
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
	GetAttemptReview(ctx context.Context, tryoutID, attemptID string) (*AttemptReviewResponse, error)
	PutAttemptAnswerReview(ctx context.Context, tryoutID, attemptID, questionID, reviewerUserID string, patch AttemptAnswerReviewPatch) (studentUserID string, newScore float64, err error)
	// AutoGradeAttempt jalankan ulang penilaian otomatis (hapus manual_score; opsional hapus komentar review).
	AutoGradeAttempt(ctx context.Context, tryoutID, attemptID string, opts AutoGradeAttemptOpts) (studentUserID string, newScore float64, err error)
}

// AutoGradeAttemptOpts POST .../auto-grade — body opsional.
type AutoGradeAttemptOpts struct {
	ClearReviewerComments bool `json:"clearReviewerComments"`
}

// AttemptAnswerReviewPatch partial update dari JSON (key hadir = ubah; manualScore null = hapus override).
type AttemptAnswerReviewPatch struct {
	HasReviewerComment bool
	ReviewerComment    *string
	HasManualScore     bool
	ManualScore        *float64 // jika HasManualScore && ManualScore == nil → kolom manual_score di DB di-NULL-kan
}

// ErrAttemptReviewNoFields body tidak memuat reviewerComment / manualScore.
var ErrAttemptReviewNoFields = errors.New("attempt review: no fields to update")

// AttemptReviewResponse GET .../attempts/{attemptId}/review — kisi jawaban untuk penilaian manual.
type AttemptReviewResponse struct {
	AttemptID   string                `json:"attemptId"`
	TryoutID    string                `json:"tryoutId"`
	Status      string                `json:"status"`
	SubmittedAt *string               `json:"submittedAt,omitempty"`
	Score       *float64              `json:"score,omitempty"`
	MaxScore    *float64              `json:"maxScore,omitempty"`
	Percentile  *float64              `json:"percentile,omitempty"`
	Student     AttemptReviewStudent  `json:"student"`
	Items       []AttemptReviewItem   `json:"items"`
}

type AttemptReviewStudent struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type AttemptReviewItem struct {
	QuestionID       string   `json:"questionId"`
	SortOrder        int      `json:"sortOrder"`
	Type             string   `json:"type"`
	Body             string   `json:"body"`
	MaxScore         float64  `json:"maxScore"`
	Options          any      `json:"options,omitempty"`
	CorrectOption    *string  `json:"correctOption,omitempty"`
	CorrectText      *string  `json:"correctText,omitempty"`
	AnswerText       *string  `json:"answerText,omitempty"`
	SelectedOption   *string  `json:"selectedOption,omitempty"`
	IsMarked         bool     `json:"isMarked"`
	AutoScore        float64  `json:"autoScore"`
	AutoIsCorrect    *bool    `json:"autoIsCorrect,omitempty"`
	ScoreGot         float64  `json:"scoreGot"`
	IsCorrect        *bool    `json:"isCorrect,omitempty"`
	ManualScore      *float64 `json:"manualScore,omitempty"`
	ReviewerComment  *string  `json:"reviewerComment,omitempty"`
	ReviewedAt       *string  `json:"reviewedAt,omitempty"`
	ReviewedByUserID *string  `json:"reviewedByUserId,omitempty"`
	ReviewedByName   *string  `json:"reviewedByName,omitempty"`
}

type MonthlyReport struct {
	Year           int   `json:"year"`
	Month          int   `json:"month"`
	NewEnrollments int   `json:"newEnrollments"`
	PaymentsCount  int   `json:"paymentsCount"`
	TotalRevenue   int64 `json:"totalRevenue"`
}

// QuestionStats response for GET /admin/tryouts/:tryoutId/questions/:questionId/stats
type QuestionStats struct {
	ParticipantsCount int     `json:"participantsCount"`
	AnsweredCount     int     `json:"answeredCount"`
	CorrectCount      int     `json:"correctCount"`
	WrongCount        int     `json:"wrongCount"`
	CorrectPercent    float64 `json:"correctPercent"`
	WrongPercent      float64 `json:"wrongPercent"`
}

// QuestionStatsItem one question in bulk stats
type QuestionStatsItem struct {
	QuestionID     string  `json:"questionId"`
	AnsweredCount  int     `json:"answeredCount"`
	CorrectCount   int     `json:"correctCount"`
	WrongCount     int     `json:"wrongCount"`
	CorrectPercent float64 `json:"correctPercent"`
	WrongPercent   float64 `json:"wrongPercent"`
}

// QuestionStatsBulk response for GET /admin/tryouts/:tryoutId/questions/stats
type QuestionStatsBulk struct {
	ParticipantsCount int                 `json:"participantsCount"`
	Questions         []QuestionStatsItem `json:"questions"`
}

// TryoutAnalysis response for GET /admin/tryouts/:tryoutId/analysis (grafik & analisis per soal)
type TryoutAnalysis struct {
	TryoutID          string                   `json:"tryoutId"`
	TryoutTitle       string                   `json:"tryoutTitle"`
	ParticipantsCount int                      `json:"participantsCount"`
	Questions         []TryoutAnalysisQuestion `json:"questions"`
}

// TryoutAnalysisQuestion satu soal dalam analisis tryout (siap untuk grafik)
type TryoutAnalysisQuestion struct {
	QuestionNumber     int                `json:"questionNumber"`
	QuestionID         string             `json:"questionId"`
	QuestionType       string             `json:"questionType"`
	AnsweredCount      int                `json:"answeredCount"`
	UnansweredCount    int                `json:"unansweredCount"`
	CorrectCount       int                `json:"correctCount"`
	WrongCount         int                `json:"wrongCount"`
	CorrectPercent     float64            `json:"correctPercent"`
	WrongPercent       float64            `json:"wrongPercent"`
	OptionDistribution map[string]int     `json:"optionDistribution"`
}

// TryoutStudentItem satu siswa yang submit tryout (untuk daftar + link ke analisis AI)
type TryoutStudentItem struct {
	UserID      string   `json:"userId"`
	UserName    string   `json:"userName"`
	UserEmail   string   `json:"userEmail"`
	AttemptID   string   `json:"attemptId"`
	Score       *float64 `json:"score"`
	MaxScore    *float64 `json:"maxScore"`
	Percentile  *float64 `json:"percentile"`
	SubmittedAt *string  `json:"submittedAt"`
}

// AttemptAIAnalysisResponse analisis AI per attempt (per siswa)
type AttemptAIAnalysisResponse struct {
	AttemptID        string   `json:"attemptId"`
	Summary          string   `json:"summary"`
	Recap            string   `json:"recap"`
	StrengthAreas    []string `json:"strengthAreas"`
	ImprovementAreas []string `json:"improvementAreas"`
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
