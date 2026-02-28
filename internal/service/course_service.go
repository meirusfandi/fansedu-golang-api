package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

type CourseService interface {
	List(ctx context.Context) ([]domain.Course, error)
	GetByID(ctx context.Context, id string) (domain.Course, error)
	Enroll(ctx context.Context, userID, courseID string) (domain.CourseEnrollment, error)
}
