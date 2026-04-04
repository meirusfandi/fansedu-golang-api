package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func CourseList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.CourseService.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.CourseResponse, len(list))
		for i := range list {
			tt := list[i].TrackType
			if tt == "" {
				tt = domain.CourseTrackMeetings
			}
			out[i] = dto.CourseResponse{
				ID:          list[i].ID,
				Title:       list[i].Title,
				Slug:        list[i].Slug,
				Description: list[i].Description,
				Price:       list[i].Price,
				Thumbnail:   list[i].Thumbnail,
				SubjectID:   list[i].SubjectID,
				CreatedBy:   list[i].CreatedBy,
				TrackType:   tt,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// CourseGetBySlug returns a single course by slug (public). GET /api/v1/courses/slug/{slug}
func CourseGetBySlug(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}
		c, err := deps.CourseRepo.GetBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, "course not found", http.StatusNotFound)
			return
		}
		tt := c.TrackType
		if tt == "" {
			tt = domain.CourseTrackMeetings
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.CourseResponse{
			ID:          c.ID,
			Title:       c.Title,
			Slug:        c.Slug,
			Description: c.Description,
			Price:       c.Price,
			Thumbnail:   c.Thumbnail,
			SubjectID:   c.SubjectID,
			CreatedBy:   c.CreatedBy,
			TrackType:   tt,
		})
	}
}

func CourseEnroll(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		courseID := chi.URLParam(r, "courseId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		enrollment, err := deps.CourseService.Enroll(r.Context(), userID, courseID)
		if err != nil {
			if err == service.ErrAlreadyEnrolled {
				http.Error(w, "already enrolled", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(dto.EnrollmentResponse{
			ID:         enrollment.ID,
			UserID:     enrollment.UserID,
			CourseID:   enrollment.CourseID,
			Status:     enrollment.Status,
			EnrolledAt: enrollment.EnrolledAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func CertificateList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		list, err := deps.CertificateRepo.ListByUserID(r.Context(), userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]map[string]interface{}, len(list))
		for i := range list {
			out[i] = map[string]interface{}{
				"id":         list[i].ID,
				"userId":   list[i].UserID,
				"issuedAt": list[i].IssuedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			if list[i].TryoutSessionID != nil {
				out[i]["tryoutSessionId"] = *list[i].TryoutSessionID
			}
			if list[i].CourseID != nil {
				out[i]["courseId"] = *list[i].CourseID
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}
