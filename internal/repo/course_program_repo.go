package repo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// ErrCourseProgramValidation input tidak valid (track, nomor pertemuan, pre-test tryout, dll.).
var ErrCourseProgramValidation = errors.New("course program validation")

// CourseProgramRepo menyimpan definisi program kelas (pertemuan, pre-test) dan menyinkronkan learning journey.
type CourseProgramRepo interface {
	GetProgram(ctx context.Context, courseID string) (track string, meetings []domain.CourseProgramMeeting, pretestSessionID *string, err error)
	// SaveProgram mengganti course_meetings, course_pretests, courses.track_type, lalu menghapus course_sections
	// untuk course ini dan membangun ulang section/lesson (lesson_progress lama untuk lesson yang terhapus ikut hilang).
	SaveProgram(ctx context.Context, courseID string, track string, meetings []domain.CourseProgramMeeting, pretestTryoutSessionID *string) error
}

type courseProgramRepo struct{ pool *pgxpool.Pool }

func NewCourseProgramRepo(pool *pgxpool.Pool) CourseProgramRepo {
	return &courseProgramRepo{pool: pool}
}

func (r *courseProgramRepo) GetProgram(ctx context.Context, courseID string) (string, []domain.CourseProgramMeeting, *string, error) {
	var track string
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(TRIM(track_type), ''), 'meetings') FROM courses WHERE id = $1::uuid
	`, courseID).Scan(&track)
	if err != nil {
		return "", nil, nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT meeting_number, title, detail_text, pdf_url, ppt_url, pr_title, pr_description, live_class_url, recording_url
		FROM course_meetings WHERE course_id = $1::uuid ORDER BY meeting_number ASC
	`, courseID)
	if err != nil {
		return "", nil, nil, err
	}
	defer rows.Close()
	var meetings []domain.CourseProgramMeeting
	for rows.Next() {
		var m domain.CourseProgramMeeting
		if err := rows.Scan(&m.MeetingNumber, &m.Title, &m.DetailText, &m.PdfURL, &m.PptURL, &m.PrTitle, &m.PrDescription, &m.LiveClassURL, &m.RecordingURL); err != nil {
			return "", nil, nil, err
		}
		meetings = append(meetings, m)
	}
	if err := rows.Err(); err != nil {
		return "", nil, nil, err
	}

	var preID string
	err = r.pool.QueryRow(ctx, `
		SELECT tryout_session_id::text FROM course_pretests WHERE course_id = $1::uuid
	`, courseID).Scan(&preID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return track, meetings, nil, nil
		}
		return "", nil, nil, err
	}
	return track, meetings, &preID, nil
}

func (r *courseProgramRepo) SaveProgram(ctx context.Context, courseID string, track string, meetings []domain.CourseProgramMeeting, pretestTryoutSessionID *string) error {
	track = strings.TrimSpace(strings.ToLower(track))
	if track == "" {
		track = domain.CourseTrackMeetings
	}
	if track != domain.CourseTrackMeetings && track != domain.CourseTrackTryout {
		return fmt.Errorf("%w: invalid track_type", ErrCourseProgramValidation)
	}

	if track == domain.CourseTrackMeetings && len(meetings) == 0 {
		return fmt.Errorf("%w: at least one meeting is required when trackType is \"meetings\"", ErrCourseProgramValidation)
	}

	seen := map[int]struct{}{}
	for _, m := range meetings {
		if m.MeetingNumber < 1 || m.MeetingNumber > 8 {
			return fmt.Errorf("%w: meeting_number must be 1..8", ErrCourseProgramValidation)
		}
		if _, ok := seen[m.MeetingNumber]; ok {
			return fmt.Errorf("%w: duplicate meeting_number %d", ErrCourseProgramValidation, m.MeetingNumber)
		}
		seen[m.MeetingNumber] = struct{}{}
	}

	if pretestTryoutSessionID != nil {
		s := strings.TrimSpace(*pretestTryoutSessionID)
		if s == "" {
			pretestTryoutSessionID = nil
		} else {
			if _, err := uuid.Parse(s); err != nil {
				return fmt.Errorf("%w: invalid pretest tryout_session_id", ErrCourseProgramValidation)
			}
			var n int
			if err := r.pool.QueryRow(ctx, `SELECT 1 FROM tryout_sessions WHERE id = $1::uuid`, s).Scan(&n); err != nil {
				if err == pgx.ErrNoRows {
					return fmt.Errorf("%w: pretest tryout session not found", ErrCourseProgramValidation)
				}
				return err
			}
			pretestTryoutSessionID = &s
		}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE courses SET track_type = $2 WHERE id = $1::uuid`, courseID, track); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM course_meetings WHERE course_id = $1::uuid`, courseID); err != nil {
		return err
	}
	for _, m := range meetings {
		if _, err := tx.Exec(ctx, `
			INSERT INTO course_meetings (course_id, meeting_number, title, detail_text, pdf_url, ppt_url, pr_title, pr_description, live_class_url, recording_url, sort_order)
			VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $2)
		`, courseID, m.MeetingNumber, m.Title, m.DetailText, m.PdfURL, m.PptURL, m.PrTitle, m.PrDescription, m.LiveClassURL, m.RecordingURL); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM course_pretests WHERE course_id = $1::uuid`, courseID); err != nil {
		return err
	}
	if pretestTryoutSessionID != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO course_pretests (course_id, tryout_session_id) VALUES ($1::uuid, $2::uuid)
		`, courseID, *pretestTryoutSessionID); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM course_sections WHERE course_id = $1::uuid`, courseID); err != nil {
		return err
	}

	sort.Slice(meetings, func(i, j int) bool { return meetings[i].MeetingNumber < meetings[j].MeetingNumber })

	switch track {
	case domain.CourseTrackMeetings:
		if err := r.rebuildJourneyMeetings(ctx, tx, courseID, meetings, pretestTryoutSessionID); err != nil {
			return err
		}
	case domain.CourseTrackTryout:
		if err := r.rebuildJourneyTryout(ctx, tx, courseID, pretestTryoutSessionID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *courseProgramRepo) rebuildJourneyMeetings(ctx context.Context, tx pgx.Tx, courseID string, meetings []domain.CourseProgramMeeting, pretest *string) error {
	secOrder := 0
	if pretest != nil && *pretest != "" {
		sid, err := insertSectionTx(ctx, tx, courseID, "Pre-test", secOrder)
		if err != nil {
			return err
		}
		secOrder++
		if err := insertLessonTx(ctx, tx, sid, "quiz", "Pre-test", 0, nil, nil, nil, nil, nil, pretest); err != nil {
			return err
		}
	}

	for _, m := range meetings {
		secTitle := fmt.Sprintf("Pertemuan %d", m.MeetingNumber)
		if t := strings.TrimSpace(m.Title); t != "" {
			secTitle = secTitle + ": " + t
		}
		sid, err := insertSectionTx(ctx, tx, courseID, secTitle, secOrder)
		if err != nil {
			return err
		}
		secOrder++

		lo := 0
		lessonTitle := strings.TrimSpace(m.Title)
		if lessonTitle == "" {
			lessonTitle = fmt.Sprintf("Pertemuan %d", m.MeetingNumber)
		}
		if err := insertLessonTx(ctx, tx, sid, "text", lessonTitle, lo, m.DetailText, nil, nil, nil, nil, nil); err != nil {
			return err
		}
		lo++

		if m.PdfURL != nil && strings.TrimSpace(*m.PdfURL) != "" {
			pdf := strings.TrimSpace(*m.PdfURL)
			if err := insertLessonTx(ctx, tx, sid, "text", "Modul PDF", lo, nil, &pdf, nil, nil, nil, nil); err != nil {
				return err
			}
			lo++
		}

		if m.PptURL != nil && strings.TrimSpace(*m.PptURL) != "" {
			ppt := strings.TrimSpace(*m.PptURL)
			if err := insertLessonTx(ctx, tx, sid, "text", "Materi PPT", lo, nil, nil, &ppt, nil, nil, nil); err != nil {
				return err
			}
			lo++
		}

		prTitle := ""
		if m.PrTitle != nil {
			prTitle = strings.TrimSpace(*m.PrTitle)
		}
		hasPR := prTitle != "" || (m.PrDescription != nil && strings.TrimSpace(*m.PrDescription) != "")
		if hasPR {
			pt := prTitle
			if pt == "" {
				pt = "Tugas (PR)"
			}
			if err := insertLessonTx(ctx, tx, sid, "assignment", pt, lo, m.PrDescription, nil, nil, nil, nil, nil); err != nil {
				return err
			}
			lo++
		}

		if m.LiveClassURL != nil && strings.TrimSpace(*m.LiveClassURL) != "" {
			live := strings.TrimSpace(*m.LiveClassURL)
			if err := insertLessonTx(ctx, tx, sid, "text", "Kelas live", lo, nil, nil, nil, &live, nil, nil); err != nil {
				return err
			}
			lo++
		}

		if m.RecordingURL != nil && strings.TrimSpace(*m.RecordingURL) != "" {
			rec := strings.TrimSpace(*m.RecordingURL)
			if err := insertLessonTx(ctx, tx, sid, "text", "Rekaman kelas", lo, nil, nil, nil, nil, &rec, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *courseProgramRepo) rebuildJourneyTryout(ctx context.Context, tx pgx.Tx, courseID string, pretest *string) error {
	secOrder := 0
	if pretest != nil && *pretest != "" {
		sid, err := insertSectionTx(ctx, tx, courseID, "Pre-test", secOrder)
		if err != nil {
			return err
		}
		secOrder++
		if err := insertLessonTx(ctx, tx, sid, "quiz", "Pre-test", 0, nil, nil, nil, nil, nil, pretest); err != nil {
			return err
		}
	}

	rows, err := tx.Query(ctx, `
		SELECT t.id::text, t.title
		FROM course_tryouts ct
		INNER JOIN tryout_sessions t ON t.id = ct.tryout_session_id
		WHERE ct.course_id = $1::uuid
		ORDER BY ct.sort_order ASC, t.title ASC
	`, courseID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sessionID, title string
		if err := rows.Scan(&sessionID, &title); err != nil {
			return err
		}
		secTitle := strings.TrimSpace(title)
		if secTitle == "" {
			secTitle = "Latihan tryout"
		}
		sid, err := insertSectionTx(ctx, tx, courseID, secTitle, secOrder)
		if err != nil {
			return err
		}
		secOrder++
		tid := sessionID
		if err := insertLessonTx(ctx, tx, sid, "quiz", secTitle, 0, nil, nil, nil, nil, nil, &tid); err != nil {
			return err
		}
	}
	return rows.Err()
}

func insertSectionTx(ctx context.Context, tx pgx.Tx, courseID, title string, sortOrder int) (string, error) {
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO course_sections (course_id, title, sort_order) VALUES ($1::uuid, $2, $3) RETURNING id::text
	`, courseID, title, sortOrder).Scan(&id)
	return id, err
}

func insertLessonTx(ctx context.Context, tx pgx.Tx, sectionID, lessonType, title string, sortOrder int, content, pdfURL, pptURL, liveURL, recordingURL, tryoutSessionID *string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO learning_lessons (section_id, type, title, sort_order, content, pdf_url, ppt_url, live_class_url, recording_url, tryout_session_id)
		VALUES ($1::uuid, $2::learning_lesson_type, $3, $4, $5, $6, $7, $8, $9, $10::uuid)
	`, sectionID, lessonType, title, sortOrder, content, pdfURL, pptURL, liveURL, recordingURL, tryoutSessionID)
	return err
}
