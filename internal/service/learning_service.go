package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
)

var (
	ErrLearningNotFound   = errors.New("learning: not found")
	ErrLearningNoAccess   = errors.New("learning: no access")
	ErrLearningLessonLock = errors.New("learning: lesson locked")
)

// LearningCourseListDTO item GET /api/v1/courses/enrolled.
type LearningCourseListDTO struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// LearningCourseJourneyDTO root GET /api/v1/courses/{ref}/journey.
type LearningCourseJourneyDTO struct {
	Course           LearningCourseMetaDTO       `json:"course"`
	ProgressPercent  float64                     `json:"progressPercent"`
	CompletedLessons int                         `json:"completedLessons"`
	TotalLessons     int                         `json:"totalLessons"`
	Sections         []LearningSectionJourneyDTO `json:"sections"`
}

type LearningCourseMetaDTO struct {
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type LearningSectionJourneyDTO struct {
	ID              string                   `json:"id"`
	CourseID        string                   `json:"courseId"`
	Title           string                   `json:"title"`
	SortOrder       int                      `json:"sortOrder"`
	ProgressPercent float64                  `json:"progressPercent"`
	Lessons         []LearningLessonBriefDTO `json:"lessons"`
}

type LearningLessonBriefDTO struct {
	ID              string  `json:"id"`
	SectionID       string  `json:"sectionId"`
	Type            string  `json:"type"`
	Title           string  `json:"title"`
	SortOrder       int     `json:"sortOrder"`
	TryoutSessionID *string `json:"tryoutSessionId,omitempty"`
	PptURL          *string `json:"pptUrl,omitempty"`
	RecordingURL    *string `json:"recordingUrl,omitempty"`
	Completed       bool    `json:"completed"`
	Locked          bool    `json:"locked"`
	ProgressPercent float64 `json:"progressPercent"`
}

// LearningLessonDetailDTO GET /api/v1/courses/lessons/:id.
type LearningLessonDetailDTO struct {
	ID              string  `json:"id"`
	SectionID       string  `json:"sectionId"`
	CourseID        string  `json:"courseId"`
	Type            string  `json:"type"`
	Title           string  `json:"title"`
	Content         *string `json:"content,omitempty"`
	PdfURL          *string `json:"pdfUrl,omitempty"`
	PptURL          *string `json:"pptUrl,omitempty"`
	LiveClassURL    *string `json:"liveClassUrl,omitempty"`
	RecordingURL    *string `json:"recordingUrl,omitempty"`
	TryoutSessionID *string `json:"tryoutSessionId,omitempty"`
	Locked          bool    `json:"locked"`
	Completed       bool    `json:"completed"`
}

// LearningLessonCompleteDTO POST .../complete.
type LearningLessonCompleteDTO struct {
	LessonID     string  `json:"lessonId"`
	CompletedAt  string  `json:"completedAt"`
	NextLessonID *string `json:"nextLessonId,omitempty"`
}

type LearningService interface {
	ListCourses(ctx context.Context, userID string) ([]LearningCourseListDTO, error)
	GetCourseJourney(ctx context.Context, userID, courseRef string) (*LearningCourseJourneyDTO, error)
	GetLesson(ctx context.Context, userID, lessonID string) (*LearningLessonDetailDTO, error)
	CompleteLesson(ctx context.Context, userID, lessonID string) (*LearningLessonCompleteDTO, error)
}

type learningService struct {
	courseRepo   repo.CourseRepo
	learningRepo repo.LearningJourneyRepo
}

func NewLearningService(courseRepo repo.CourseRepo, learningRepo repo.LearningJourneyRepo) LearningService {
	return &learningService{courseRepo: courseRepo, learningRepo: learningRepo}
}

func (s *learningService) resolveCourseID(ctx context.Context, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", pgx.ErrNoRows
	}
	if _, err := uuid.Parse(ref); err == nil {
		if _, err := s.courseRepo.GetByID(ctx, ref); err != nil {
			return "", err
		}
		return ref, nil
	}
	c, err := s.courseRepo.GetBySlug(ctx, ref)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

func courseSlugString(c domain.Course) string {
	if c.Slug != nil && strings.TrimSpace(*c.Slug) != "" {
		return strings.TrimSpace(*c.Slug)
	}
	return ""
}

func (s *learningService) ListCourses(ctx context.Context, userID string) ([]LearningCourseListDTO, error) {
	rows, err := s.learningRepo.ListEnrolledCourses(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]LearningCourseListDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, LearningCourseListDTO{ID: r.ID, Slug: r.Slug, Title: r.Title})
	}
	return out, nil
}

func (s *learningService) GetCourseJourney(ctx context.Context, userID, courseRef string) (*LearningCourseJourneyDTO, error) {
	courseID, err := s.resolveCourseID(ctx, courseRef)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLearningNotFound
		}
		return nil, ErrLearningNotFound
	}
	ok, err := s.learningRepo.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLearningNoAccess
	}
	c, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return nil, ErrLearningNotFound
	}
	sections, err := s.learningRepo.ListSectionsForCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	orderedLessons, err := s.learningRepo.ListLessonRowsForCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	done, err := s.learningRepo.ListCompletedLessonIDsForCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}

	total := len(orderedLessons)
	completedN := 0
	for _, l := range orderedLessons {
		if _, ok := done[l.ID]; ok {
			completedN++
		}
	}
	var courseProgress float64
	if total > 0 {
		courseProgress = math.Round(float64(completedN)/float64(total)*10000) / 100
	}

	// locked by global order: first lesson unlocked; lesson i locked iff previous not completed
	lockedByID := make(map[string]bool, len(orderedLessons))
	for i := range orderedLessons {
		if i == 0 {
			lockedByID[orderedLessons[i].ID] = false
			continue
		}
		prevID := orderedLessons[i-1].ID
		_, prevDone := done[prevID]
		lockedByID[orderedLessons[i].ID] = !prevDone
	}

	lessonsBySection := make(map[string][]repo.LearningLessonRow)
	for _, l := range orderedLessons {
		lessonsBySection[l.SectionID] = append(lessonsBySection[l.SectionID], l)
	}

	desc := c.Description
	slugStr := courseSlugString(c)
	meta := LearningCourseMetaDTO{ID: c.ID, Slug: slugStr, Title: c.Title, Description: desc}

	secDTOs := make([]LearningSectionJourneyDTO, 0, len(sections))
	for _, sec := range sections {
		ls := lessonsBySection[sec.ID]
		secDone := 0
		for _, l := range ls {
			if _, ok := done[l.ID]; ok {
				secDone++
			}
		}
		secTotal := len(ls)
		var secPct float64
		if secTotal > 0 {
			secPct = math.Round(float64(secDone)/float64(secTotal)*10000) / 100
		}
		lessBrief := make([]LearningLessonBriefDTO, 0, len(ls))
		for _, l := range ls {
			_, comp := done[l.ID]
			lp := 0.0
			if comp {
				lp = 100
			}
			lessBrief = append(lessBrief, LearningLessonBriefDTO{
				ID:              l.ID,
				SectionID:       l.SectionID,
				Type:            l.Type,
				Title:           l.Title,
				SortOrder:       l.SortOrder,
				TryoutSessionID: l.TryoutSessionID,
				PptURL:          l.PptURL,
				RecordingURL:    l.RecordingURL,
				Completed:       comp,
				Locked:          lockedByID[l.ID],
				ProgressPercent: lp,
			})
		}
		secDTOs = append(secDTOs, LearningSectionJourneyDTO{
			ID:              sec.ID,
			CourseID:        sec.CourseID,
			Title:           sec.Title,
			SortOrder:       sec.SortOrder,
			ProgressPercent: secPct,
			Lessons:         lessBrief,
		})
	}

	return &LearningCourseJourneyDTO{
		Course:           meta,
		ProgressPercent:  courseProgress,
		CompletedLessons: completedN,
		TotalLessons:     total,
		Sections:         secDTOs,
	}, nil
}

func (s *learningService) GetLesson(ctx context.Context, userID, lessonID string) (*LearningLessonDetailDTO, error) {
	if _, err := uuid.Parse(strings.TrimSpace(lessonID)); err != nil {
		return nil, ErrLearningNotFound
	}
	row, courseID, _, err := s.learningRepo.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLearningNotFound
		}
		return nil, err
	}
	ok, err := s.learningRepo.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLearningNoAccess
	}
	orderedLessons, err := s.learningRepo.ListLessonRowsForCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	done, err := s.learningRepo.ListCompletedLessonIDsForCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	locked := false
	for i := range orderedLessons {
		if orderedLessons[i].ID != lessonID {
			continue
		}
		if i == 0 {
			locked = false
		} else {
			_, prevDone := done[orderedLessons[i-1].ID]
			locked = !prevDone
		}
		break
	}
	_, comp := done[lessonID]
	return &LearningLessonDetailDTO{
		ID:              row.ID,
		SectionID:       row.SectionID,
		CourseID:        courseID,
		Type:            row.Type,
		Title:           row.Title,
		Content:         row.Content,
		PdfURL:          row.PdfURL,
		PptURL:          row.PptURL,
		LiveClassURL:    row.LiveClassURL,
		RecordingURL:    row.RecordingURL,
		TryoutSessionID: row.TryoutSessionID,
		Locked:          locked,
		Completed:       comp,
	}, nil
}

func (s *learningService) CompleteLesson(ctx context.Context, userID, lessonID string) (*LearningLessonCompleteDTO, error) {
	if _, err := uuid.Parse(strings.TrimSpace(lessonID)); err != nil {
		return nil, ErrLearningNotFound
	}
	_, courseID, _, err := s.learningRepo.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLearningNotFound
		}
		return nil, err
	}
	ok, err := s.learningRepo.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLearningNoAccess
	}
	orderedLessons, err := s.learningRepo.ListLessonRowsForCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	done, err := s.learningRepo.ListCompletedLessonIDsForCourse(ctx, userID, courseID)
	if err != nil {
		return nil, err
	}
	idx := -1
	for i := range orderedLessons {
		if orderedLessons[i].ID == lessonID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, ErrLearningNotFound
	}
	if idx > 0 {
		_, prevDone := done[orderedLessons[idx-1].ID]
		if !prevDone {
			return nil, ErrLearningLessonLock
		}
	}
	at, err := s.learningRepo.UpsertLessonProgress(ctx, userID, lessonID)
	if err != nil {
		return nil, err
	}
	var nextID *string
	if idx+1 < len(orderedLessons) {
		n := orderedLessons[idx+1].ID
		nextID = &n
	}
	atStr := at.UTC().Format(time.RFC3339)
	return &LearningLessonCompleteDTO{
		LessonID:     lessonID,
		CompletedAt:  atStr,
		NextLessonID: nextID,
	}, nil
}
