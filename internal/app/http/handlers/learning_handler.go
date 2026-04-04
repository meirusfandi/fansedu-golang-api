package handlers

// Learning journey dipasang di /api/v1/courses (bukan prefix /learning):
//   GET  /courses/enrolled
//   GET  /courses/{courseRef}/journey   — courseRef = UUID atau slug
//   GET  /courses/lessons/{lessonId}
//   POST /courses/lessons/{lessonId}/complete
// Semua butuh Auth + PasswordSetupGuard (sama seperti /student).

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// LearningCoursesList GET /api/v1/courses/enrolled
func LearningCoursesList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.LearningService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan learning journey belum tersedia.")
			return
		}
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		list, err := deps.LearningService.ListCourses(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": list})
	}
}

// LearningCourseGet GET /api/v1/courses/{courseRef}/journey
func LearningCourseGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.LearningService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan learning journey belum tersedia.")
			return
		}
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		ref := chi.URLParam(r, "courseRef")
		out, err := deps.LearningService.GetCourseJourney(r.Context(), userID, ref)
		if err != nil {
			if errors.Is(err, service.ErrLearningNotFound) || errors.Is(err, service.ErrLearningNoAccess) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Kursus tidak ditemukan atau Anda tidak memiliki akses.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

// LearningLessonGet GET /api/v1/courses/lessons/{lessonId}
func LearningLessonGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.LearningService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan learning journey belum tersedia.")
			return
		}
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		lessonID := chi.URLParam(r, "lessonId")
		out, err := deps.LearningService.GetLesson(r.Context(), userID, lessonID)
		if err != nil {
			if errors.Is(err, service.ErrLearningNotFound) || errors.Is(err, service.ErrLearningNoAccess) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Materi tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}

// LearningLessonComplete POST /api/v1/courses/lessons/{lessonId}/complete
func LearningLessonComplete(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.LearningService == nil {
			writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Layanan learning journey belum tersedia.")
			return
		}
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		lessonID := chi.URLParam(r, "lessonId")
		out, err := deps.LearningService.CompleteLesson(r.Context(), userID, lessonID)
		if err != nil {
			if errors.Is(err, service.ErrLearningLessonLock) {
				writeError(w, http.StatusForbidden, "LESSON_LOCKED", "Selesaikan materi sebelumnya terlebih dahulu.")
				return
			}
			if errors.Is(err, service.ErrLearningNotFound) || errors.Is(err, service.ErrLearningNoAccess) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Materi tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
	}
}
