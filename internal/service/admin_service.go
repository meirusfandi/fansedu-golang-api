package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type AdminOverview struct {
	TotalStudents   int     `json:"total_students"`
	ActiveTryouts   int     `json:"active_tryouts"`
	AvgScore        float64 `json:"avg_score"`
	TotalCertificates int   `json:"total_certificates"`
}

type AdminService interface {
	Overview(ctx context.Context) (*AdminOverview, error)
	CreateTryout(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
	UpdateTryout(ctx context.Context, t domain.TryoutSession) error
	DeleteTryout(ctx context.Context, id string) error
	CreateQuestion(ctx context.Context, q domain.Question) (domain.Question, error)
	UpdateQuestion(ctx context.Context, q domain.Question) error
	DeleteQuestion(ctx context.Context, id string) error
	CreateCourse(ctx context.Context, c domain.Course) (domain.Course, error)
	ListEnrollmentsByCourse(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	IssueCertificate(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
}
