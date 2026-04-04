package service

import (
	"context"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

// CourseProgramService membaca/menulis format program kelas admin (pertemuan, pre-test, track) dan sync journey.
type CourseProgramService interface {
	GetProgram(ctx context.Context, courseID string) (track string, meetings []domain.CourseProgramMeeting, pretestSessionID *string, err error)
	SaveProgram(ctx context.Context, courseID string, track string, meetings []domain.CourseProgramMeeting, pretestSessionID *string) error
}

type courseProgramService struct {
	courses repo.CourseRepo
	prog    repo.CourseProgramRepo
}

func NewCourseProgramService(courses repo.CourseRepo, prog repo.CourseProgramRepo) CourseProgramService {
	return &courseProgramService{courses: courses, prog: prog}
}

func (s *courseProgramService) GetProgram(ctx context.Context, courseID string) (string, []domain.CourseProgramMeeting, *string, error) {
	if _, err := s.courses.GetByID(ctx, courseID); err != nil {
		return "", nil, nil, err
	}
	return s.prog.GetProgram(ctx, courseID)
}

func (s *courseProgramService) SaveProgram(ctx context.Context, courseID string, track string, meetings []domain.CourseProgramMeeting, pretestSessionID *string) error {
	if _, err := s.courses.GetByID(ctx, courseID); err != nil {
		return err
	}
	return s.prog.SaveProgram(ctx, courseID, track, meetings, pretestSessionID)
}
