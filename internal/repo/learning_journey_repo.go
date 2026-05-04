package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LearningCourseRow ringkas untuk GET /api/v1/courses/enrolled.
type LearningCourseRow struct {
	ID    string
	Slug  string
	Title string
}

// LearningSectionRow + LearningLessonRow untuk query tree.
type LearningSectionRow struct {
	ID        string
	CourseID  string
	Title     string
	SortOrder int
}

type LearningLessonRow struct {
	ID              string
	SectionID       string
	Type            string
	Title           string
	SortOrder       int
	Content         *string
	PdfURL          *string
	PptURL          *string
	LiveClassURL    *string
	RecordingURL    *string
	TryoutSessionID *string
}

type LearningJourneyRepo interface {
	ListEnrolledCourses(ctx context.Context, userID string) ([]LearningCourseRow, error)
	IsEnrolled(ctx context.Context, userID, courseID string) (bool, error)
	ListSectionsForCourse(ctx context.Context, courseID string) ([]LearningSectionRow, error)
	ListLessonsForSection(ctx context.Context, sectionID string) ([]LearningLessonRow, error)
	ListLessonRowsForCourse(ctx context.Context, courseID string) ([]LearningLessonRow, error)
	GetLessonByID(ctx context.Context, lessonID string) (LearningLessonRow, string, string, error) // lesson, courseID, sectionID
	ListCompletedLessonIDsForCourse(ctx context.Context, userID, courseID string) (map[string]time.Time, error)
	UpsertLessonProgress(ctx context.Context, userID, lessonID string) (completedAt time.Time, err error)
	GetLessonProgress(ctx context.Context, userID, lessonID string) (completedAt time.Time, ok bool, err error)
}

type learningJourneyRepo struct{ pool *pgxpool.Pool }

func NewLearningJourneyRepo(pool *pgxpool.Pool) LearningJourneyRepo {
	return &learningJourneyRepo{pool: pool}
}

func (r *learningJourneyRepo) ListEnrolledCourses(ctx context.Context, userID string) ([]LearningCourseRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id::text, COALESCE(c.slug, ''), c.title
		FROM courses c
		INNER JOIN course_enrollments e ON e.course_id = c.id AND e.user_id = $1::uuid
		ORDER BY e.enrolled_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LearningCourseRow
	for rows.Next() {
		var row LearningCourseRow
		if err := rows.Scan(&row.ID, &row.Slug, &row.Title); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *learningJourneyRepo) IsEnrolled(ctx context.Context, userID, courseID string) (bool, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT 1 FROM course_enrollments WHERE user_id = $1::uuid AND course_id = $2::uuid LIMIT 1
	`, userID, courseID).Scan(&n)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *learningJourneyRepo) ListSectionsForCourse(ctx context.Context, courseID string) ([]LearningSectionRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, course_id::text, title, sort_order
		FROM course_sections WHERE course_id = $1::uuid
		ORDER BY sort_order ASC, id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []LearningSectionRow
	for rows.Next() {
		var s LearningSectionRow
		if err := rows.Scan(&s.ID, &s.CourseID, &s.Title, &s.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *learningJourneyRepo) ListLessonsForSection(ctx context.Context, sectionID string) ([]LearningLessonRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, section_id::text, type::text, title, sort_order, content, pdf_url, ppt_url, live_class_url, recording_url, tryout_session_id::text
		FROM learning_lessons WHERE section_id = $1::uuid
		ORDER BY sort_order ASC, id ASC
	`, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLearningLessonRows(rows)
}

func (r *learningJourneyRepo) ListLessonRowsForCourse(ctx context.Context, courseID string) ([]LearningLessonRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT l.id::text, l.section_id::text, l.type::text, l.title, l.sort_order, l.content, l.pdf_url, l.ppt_url, l.live_class_url, l.recording_url, l.tryout_session_id::text
		FROM learning_lessons l
		INNER JOIN course_sections s ON s.id = l.section_id
		WHERE s.course_id = $1::uuid
		ORDER BY s.sort_order ASC, s.id ASC, l.sort_order ASC, l.id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLearningLessonRows(rows)
}

func scanLearningLessonRows(rows pgx.Rows) ([]LearningLessonRow, error) {
	var list []LearningLessonRow
	for rows.Next() {
		var l LearningLessonRow
		if err := rows.Scan(&l.ID, &l.SectionID, &l.Type, &l.Title, &l.SortOrder, &l.Content, &l.PdfURL, &l.PptURL, &l.LiveClassURL, &l.RecordingURL, &l.TryoutSessionID); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

func (r *learningJourneyRepo) GetLessonByID(ctx context.Context, lessonID string) (LearningLessonRow, string, string, error) {
	var l LearningLessonRow
	var courseID string
	err := r.pool.QueryRow(ctx, `
		SELECT l.id::text, l.section_id::text, l.type::text, l.title, l.sort_order, l.content, l.pdf_url, l.ppt_url, l.live_class_url, l.recording_url, l.tryout_session_id::text,
		       s.course_id::text
		FROM learning_lessons l
		INNER JOIN course_sections s ON s.id = l.section_id
		WHERE l.id = $1::uuid
	`, lessonID).Scan(
		&l.ID, &l.SectionID, &l.Type, &l.Title, &l.SortOrder, &l.Content, &l.PdfURL, &l.PptURL, &l.LiveClassURL, &l.RecordingURL, &l.TryoutSessionID,
		&courseID,
	)
	return l, courseID, l.SectionID, err
}

func (r *learningJourneyRepo) ListCompletedLessonIDsForCourse(ctx context.Context, userID, courseID string) (map[string]time.Time, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT lp.lesson_id::text, lp.completed_at
		FROM lesson_progress lp
		INNER JOIN learning_lessons l ON l.id = lp.lesson_id
		INNER JOIN course_sections s ON s.id = l.section_id
		WHERE lp.user_id = $1::uuid AND s.course_id = $2::uuid
	`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]time.Time)
	for rows.Next() {
		var lid string
		var at time.Time
		if err := rows.Scan(&lid, &at); err != nil {
			return nil, err
		}
		m[lid] = at
	}
	return m, rows.Err()
}

func (r *learningJourneyRepo) UpsertLessonProgress(ctx context.Context, userID, lessonID string) (time.Time, error) {
	var at time.Time
	err := r.pool.QueryRow(ctx, `
		INSERT INTO lesson_progress (user_id, lesson_id, completed_at)
		VALUES ($1::uuid, $2::uuid, NOW())
		ON CONFLICT (user_id, lesson_id) DO UPDATE SET completed_at = lesson_progress.completed_at
		RETURNING completed_at
	`, userID, lessonID).Scan(&at)
	return at, err
}

func (r *learningJourneyRepo) GetLessonProgress(ctx context.Context, userID, lessonID string) (time.Time, bool, error) {
	var at time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT completed_at FROM lesson_progress WHERE user_id = $1::uuid AND lesson_id = $2::uuid
	`, userID, lessonID).Scan(&at)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, false, nil
		}
		return time.Time{}, false, err
	}
	return at, true, nil
}
