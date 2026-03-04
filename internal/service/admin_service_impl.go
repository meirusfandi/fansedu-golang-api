package service

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type adminService struct {
	userRepo interface {
		List(ctx context.Context, role string) ([]domain.User, error)
		FindByID(ctx context.Context, id string) (domain.User, error)
		Create(ctx context.Context, u domain.User) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
		Count(ctx context.Context) (int, error)
	}
	tryoutRepo interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		List(ctx context.Context) ([]domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	}
	questionRepo interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		GetByID(ctx context.Context, id string) (domain.Question, error)
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	}
	courseRepo interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
		List(ctx context.Context) ([]domain.Course, error)
		Update(ctx context.Context, c domain.Course) error
		Count(ctx context.Context) (int, error)
	}
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
		Count(ctx context.Context) (int, error)
		CountEnrolledInMonth(ctx context.Context, year, month int) (int, error)
	}
	courseContentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error)
		GetByID(ctx context.Context, id string) (domain.CourseContent, error)
		Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
		Update(ctx context.Context, c domain.CourseContent) error
		Delete(ctx context.Context, id string) error
	}
	paymentRepo interface {
		Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
		List(ctx context.Context, limit int) ([]domain.Payment, error)
		CountPaidInMonth(ctx context.Context, year, month int) (int, error)
		TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
	}
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	}
	userCount  func(ctx context.Context) (int, error)
	attemptAvg func(ctx context.Context) (float64, error)
	certCount  func(ctx context.Context) (int, error)
}

func NewAdminService(
	userRepo interface {
		List(ctx context.Context, role string) ([]domain.User, error)
		FindByID(ctx context.Context, id string) (domain.User, error)
		Create(ctx context.Context, u domain.User) (domain.User, error)
		Update(ctx context.Context, u domain.User) error
		Count(ctx context.Context) (int, error)
	},
	tryoutRepo interface {
		Create(ctx context.Context, t domain.TryoutSession) (domain.TryoutSession, error)
		GetByID(ctx context.Context, id string) (domain.TryoutSession, error)
		List(ctx context.Context) ([]domain.TryoutSession, error)
		ListOpen(ctx context.Context, now time.Time) ([]domain.TryoutSession, error)
		Update(ctx context.Context, t domain.TryoutSession) error
		Delete(ctx context.Context, id string) error
	},
	questionRepo interface {
		Create(ctx context.Context, q domain.Question) (domain.Question, error)
		GetByID(ctx context.Context, id string) (domain.Question, error)
		ListByTryoutSessionID(ctx context.Context, tryoutSessionID string) ([]domain.Question, error)
		Update(ctx context.Context, q domain.Question) error
		Delete(ctx context.Context, id string) error
	},
	courseRepo interface {
		Create(ctx context.Context, c domain.Course) (domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
		List(ctx context.Context) ([]domain.Course, error)
		Update(ctx context.Context, c domain.Course) error
		Count(ctx context.Context) (int, error)
	},
	enrollmentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseEnrollment, error)
		Count(ctx context.Context) (int, error)
		CountEnrolledInMonth(ctx context.Context, year, month int) (int, error)
	},
	courseContentRepo interface {
		ListByCourseID(ctx context.Context, courseID string) ([]domain.CourseContent, error)
		GetByID(ctx context.Context, id string) (domain.CourseContent, error)
		Create(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error)
		Update(ctx context.Context, c domain.CourseContent) error
		Delete(ctx context.Context, id string) error
	},
	paymentRepo interface {
		Create(ctx context.Context, p domain.Payment) (domain.Payment, error)
		List(ctx context.Context, limit int) ([]domain.Payment, error)
		CountPaidInMonth(ctx context.Context, year, month int) (int, error)
		TotalAmountPaidInMonth(ctx context.Context, year, month int) (int64, error)
	},
	certificateRepo interface {
		Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error)
	},
	userCount func(ctx context.Context) (int, error),
	attemptAvg func(ctx context.Context) (float64, error),
	certCount func(ctx context.Context) (int, error),
) AdminService {
	return &adminService{
		userRepo:         userRepo,
		tryoutRepo:       tryoutRepo,
		questionRepo:     questionRepo,
		courseRepo:       courseRepo,
		enrollmentRepo:   enrollmentRepo,
		courseContentRepo: courseContentRepo,
		paymentRepo:      paymentRepo,
		certificateRepo:  certificateRepo,
		userCount:        userCount,
		attemptAvg:       attemptAvg,
		certCount:        certCount,
	}
}

func (s *adminService) Overview(ctx context.Context) (*AdminOverview, error) {
	students, _ := s.userCount(ctx)
	totalUsers, _ := s.userRepo.Count(ctx)
	open, _ := s.tryoutRepo.ListOpen(ctx, time.Now())
	totalCourses, _ := s.courseRepo.Count(ctx)
	totalEnrollments, _ := s.enrollmentRepo.Count(ctx)
	avg, _ := s.attemptAvg(ctx)
	certs, _ := s.certCount(ctx)
	return &AdminOverview{
		TotalStudents:     students,
		TotalUsers:        totalUsers,
		ActiveTryouts:     len(open),
		TotalCourses:      totalCourses,
		TotalEnrollments:  totalEnrollments,
		AvgScore:          avg,
		TotalCertificates: certs,
	}, nil
}

func (s *adminService) ListUsers(ctx context.Context, role string) ([]domain.User, error) {
	return s.userRepo.List(ctx, role)
}

func (s *adminService) GetUser(ctx context.Context, id string) (domain.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *adminService) CreateUser(ctx context.Context, u domain.User, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	u.PasswordHash = string(hash)
	if u.Role == "" {
		u.Role = domain.UserRoleStudent
	}
	return s.userRepo.Create(ctx, u)
}

func (s *adminService) UpdateUser(ctx context.Context, u domain.User, newPassword *string) error {
	if newPassword != nil && *newPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*newPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hash)
	}
	return s.userRepo.Update(ctx, u)
}

func (s *adminService) ListCourses(ctx context.Context) ([]domain.Course, error) {
	return s.courseRepo.List(ctx)
}

func (s *adminService) GetCourseByID(ctx context.Context, id string) (domain.Course, error) {
	return s.courseRepo.GetByID(ctx, id)
}

func (s *adminService) UpdateCourse(ctx context.Context, c domain.Course) error {
	return s.courseRepo.Update(ctx, c)
}

func (s *adminService) ListTryouts(ctx context.Context) ([]domain.TryoutSession, error) {
	return s.tryoutRepo.List(ctx)
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

func (s *adminService) ListQuestions(ctx context.Context, tryoutID string) ([]domain.Question, error) {
	return s.questionRepo.ListByTryoutSessionID(ctx, tryoutID)
}

func (s *adminService) GetQuestion(ctx context.Context, questionID string) (domain.Question, error) {
	return s.questionRepo.GetByID(ctx, questionID)
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

func (s *adminService) ListCourseContents(ctx context.Context, courseID string) ([]domain.CourseContent, error) {
	return s.courseContentRepo.ListByCourseID(ctx, courseID)
}

func (s *adminService) GetCourseContent(ctx context.Context, id string) (domain.CourseContent, error) {
	return s.courseContentRepo.GetByID(ctx, id)
}

func (s *adminService) CreateCourseContent(ctx context.Context, c domain.CourseContent) (domain.CourseContent, error) {
	return s.courseContentRepo.Create(ctx, c)
}

func (s *adminService) UpdateCourseContent(ctx context.Context, c domain.CourseContent) error {
	return s.courseContentRepo.Update(ctx, c)
}

func (s *adminService) DeleteCourseContent(ctx context.Context, id string) error {
	return s.courseContentRepo.Delete(ctx, id)
}

func (s *adminService) ListPayments(ctx context.Context, limit int) ([]domain.Payment, error) {
	return s.paymentRepo.List(ctx, limit)
}

func (s *adminService) CreatePayment(ctx context.Context, p domain.Payment) (domain.Payment, error) {
	return s.paymentRepo.Create(ctx, p)
}

func (s *adminService) ReportMonthly(ctx context.Context, year, month int) (*MonthlyReport, error) {
	enrollments, _ := s.enrollmentRepo.CountEnrolledInMonth(ctx, year, month)
	paymentsCount, _ := s.paymentRepo.CountPaidInMonth(ctx, year, month)
	revenue, _ := s.paymentRepo.TotalAmountPaidInMonth(ctx, year, month)
	return &MonthlyReport{
		Year:              year,
		Month:             month,
		NewEnrollments:    enrollments,
		PaymentsCount:     paymentsCount,
		TotalRevenueCents: revenue,
	}, nil
}
