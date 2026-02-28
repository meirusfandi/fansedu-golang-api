package service

import (
	"context"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type adminService struct {
	userRepo       interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
	}
	tryoutRepo     interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	}
	questionRepo   interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	}
	courseRepo     interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
	}
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	}
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	}
	userCount  func(ctx context.Context) (int, error)
	attemptAvg func(ctx context.Context) (float64, error)
	certCount  func(ctx context.Context) (int, error)
}

func NewAdminService(
	tryoutRepo interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	},
	questionRepo interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	},
	courseRepo interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
	},
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
	},
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	},
	userCount func(ctx context.Context) (int, error),
	attemptAvg func(ctx context.Context) (float64, error),
	certCount func(ctx context.Context) (int, error),
) AdminService {
	return &adminService{
		tryoutRepo:     tryoutRepo,
		questionRepo:   questionRepo,
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
		certificateRepo: certificateRepo,
		userCount:      userCount,
		attemptAvg:     attemptAvg,
		certCount:      certCount,
	}
}

func (s *adminService) Overview(ctx context.Context) (*AdminOverview, error) {
	students, _ := s.userCount(ctx)
	open, _ := s.tryoutRepo.ListOpen(ctx, time.Now())
	avg, _ := s.attemptAvg(ctx)
	certs, _ := s.certCount(ctx)
	return &AdminOverview{
		TotalStudents:     students,
		ActiveTryouts:     len(open),
		AvgScore:          avg,
		TotalCertificates: certs,
	}, nil
}

func (s *adminService) CreateTryout(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error) {
	return s.tryoutRepo.Create(ctx, t)
}

func (s *adminService) UpdateTryout(ctx context.Context, t domain.TryoutSession) error {
	return s.tryoutRepo.Update(ctx, t)
}

func (s *adminService) DeleteTryout(ctx context.Context, id string) error {
	return s.tryoutRepo.Delete(ctx, id)
}

func (s *adminService) CreateQuestion(ctx context.Context, q domain.Question) (domain.Question, error) {
	return s.questionRepo.Create(ctx, q)
}

func (s *adminService) UpdateQuestion(ctx context.Context, q domain.Question) error {
	return s.questionRepo.Update(ctx, q)
}

func (s *adminService) DeleteQuestion(ctx context.Context, id string) error {
	return s.questionRepo.Delete(ctx, id)
}

func (s *adminService) CreateCourse(ctx context.Context, c domain.Course) (domain.Course, error) {
	return s.courseRepo.Create(ctx, c)
}

func (s *adminService) ListEnrollmentsByCourse(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error) {
	return s.enrollmentRepo.ListByCourseID(ctx, courseID)
}

func (s *adminService) IssueCertificate(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	return s.certificateRepo.Create(ctx, c)
}
