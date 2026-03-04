package service

import (
	"context"

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
	ReportMonthly(ctx context.Context, year, month int) (*MonthlyReport, error)
}

type MonthlyReport struct {
	Year               int   `json:"year"`
	Month              int   `json:"month"`
	NewEnrollments     int   `json:"new_enrollments"`
	PaymentsCount      int   `json:"payments_count"`
	TotalRevenueCents  int64 `json:"total_revenue_cents"`
}
