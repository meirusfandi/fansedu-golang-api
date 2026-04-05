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
	"github.com/meirusfandi/fansedu-golang-api/internal/repo"
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
				MeetingNumber:  m.MeetingNumber,
				Title:          m.Title,
				DetailText:     m.DetailText,
				PdfURL:         m.PdfURL,
				PptURL:         m.PptURL,
				PrTitle:        m.PrTitle,
				PrDescription:  m.PrDescription,
				LiveClassURL:   m.LiveClassURL,
			})
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": dto.AdminCourseProgramResponse{
				TrackType:               track,
				Meetings:                items,
				PretestTryoutSessionID:  pre,
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
		var req dto.AdminCourseProgramPutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid body")
			return
		}
		meetings := make([]domain.CourseProgramMeeting, 0, len(req.Meetings))
		for _, it := range req.Meetings {
			meetings = append(meetings, domain.CourseProgramMeeting{
				MeetingNumber:  it.MeetingNumber,
				Title:          it.Title,
				DetailText:     it.DetailText,
				PdfURL:         it.PdfURL,
				PptURL:         it.PptURL,
				PrTitle:        it.PrTitle,
				PrDescription:  it.PrDescription,
				LiveClassURL:   it.LiveClassURL,
			})
		}
		track := strings.TrimSpace(strings.ToLower(req.TrackType))
		if track == "" {
			track = domain.CourseTrackMeetings
		} else if track != domain.CourseTrackMeetings && track != domain.CourseTrackTryout {
			writeError(w, http.StatusBadRequest, "validation_error", "trackType must be \"meetings\" or \"tryout\"")
			return
		}
		err = deps.CourseProgramService.SaveProgram(r.Context(), courseID, track, meetings, req.PretestTryoutSessionID)
		if err != nil {
			if errors.Is(err, repo.ErrCourseProgramValidation) {
				writeError(w, http.StatusBadRequest, "validation_error", err.Error())
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "course program saved; learning journey rebuilt"})
	}
}
