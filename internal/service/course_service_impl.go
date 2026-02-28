package service

import (
	"context"
	"errors"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

var ErrAlreadyEnrolled = errors.New("already enrolled")

type courseService struct {
	courseRepo     interface {
		List(ctx context.Context) ([]domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
	}
	enrollmentRepo interface {
		Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
		GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	}
}

func NewCourseService(
	courseRepo interface {
		List(ctx context.Context) ([]domain.Course, error)
		GetByID(ctx context.Context, id string) (domain.Course, error)
	},
	enrollmentRepo interface {
		Create(ctx context.Context, e domain.CourseEnrollment) (domain.CourseEnrollment, error)
		GetByUserAndCourse(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
	},
) CourseService {
	return &courseService{
		courseRepo:     courseRepo,
		enrollmentRepo: enrollmentRepo,
	}
}

func (s *courseService) List(ctx context.Context) ([]domain.Course, error) {
	return s.courseRepo.List(ctx)
}

func (s *courseService) GetByID(ctx context.Context, id string) (domain.Course, error) {
	return s.courseRepo.GetByID(ctx, id)
}

func (s *courseService) Enroll(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error) {
	_, err := s.enrollmentRepo.GetByUserAndCourse(ctx, userID, courseID)
	if err == nil {
		return domain.CourseEnrollment{}, ErrAlreadyEnrolled
	}
	e := domain.CourseEnrollment{
		UserID:   userID,
		CourseID: courseID,
		Status:   domain.EnrollmentStatusEnrolled,
	}
	return s.enrollmentRepo.Create(ctx, e)
}
