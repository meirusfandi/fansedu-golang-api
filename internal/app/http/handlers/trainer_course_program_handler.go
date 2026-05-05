package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func trainerOwnsCourse(c domain.Course, trainerUserID string) bool {
	if c.CreatedBy == nil {
		return false
	}
	return strings.TrimSpace(*c.CreatedBy) == strings.TrimSpace(trainerUserID)
}

// TrainerCourseProgramGet GET /api/v1/trainer/courses/{courseId}/program — hanya kelas yang created_by = trainer.
func TrainerCourseProgramGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.CourseProgramService == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course program service not configured")
			return
		}
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		courseID := strings.TrimSpace(chi.URLParam(r, "courseId"))
		if courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId required")
			return
		}
		if _, err := uuid.Parse(courseID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid courseId")
			return
		}
		c, err := deps.CourseRepo.GetByID(r.Context(), courseID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found", "course not found")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if !trainerOwnsCourse(c, userID) {
			writeError(w, http.StatusForbidden, "forbidden", "you can only manage your own courses")
			return
		}
		track, meetings, pre, err := deps.CourseProgramService.GetProgram(r.Context(), courseID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found", "course not found")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		items := make([]dto.AdminCourseProgramMeetingItem, 0, len(meetings))
		for _, m := range meetings {
			items = append(items, dto.AdminCourseProgramMeetingItem{
				MeetingNumber: m.MeetingNumber,
				Title:         m.Title,
				DetailText:    m.DetailText,
				PdfURL:        m.PdfURL,
				PptURL:        m.PptURL,
				PrTitle:       m.PrTitle,
				PrDescription: m.PrDescription,
				LiveClassURL:  m.LiveClassURL,
				RecordingURL:  m.RecordingURL,
			})
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": dto.AdminCourseProgramResponse{
				TrackType:              track,
				Meetings:               items,
				PretestTryoutSessionID: pre,
			},
		})
	}
}

// TrainerCourseProgramPut PUT /api/v1/trainer/courses/{courseId}/program
func TrainerCourseProgramPut(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.CourseProgramService == nil {
			writeError(w, http.StatusServiceUnavailable, "service_unavailable", "course program service not configured")
			return
		}
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}
		courseID := strings.TrimSpace(chi.URLParam(r, "courseId"))
		if courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId required")
			return
		}
		if _, err := uuid.Parse(courseID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid courseId")
			return
		}
		c, err := deps.CourseRepo.GetByID(r.Context(), courseID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found", "course not found")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if !trainerOwnsCourse(c, userID) {
			writeError(w, http.StatusForbidden, "forbidden", "you can only manage your own courses")
			return
		}
		writeError(w, http.StatusForbidden, "forbidden", "program and learning journey can only be saved via admin LMS (PUT /api/v1/admin/courses/{courseId}/program)")
	}
}
