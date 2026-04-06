package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// AdminGrantEnrollment POST /api/v1/admin/enrollments/grant — beri akses kelas ke user tanpa order.
func AdminGrantEnrollment(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AdminGrantEnrollmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		userID := strings.TrimSpace(req.UserID)
		courseID := strings.TrimSpace(req.CourseID)
		if userID == "" || courseID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "userId dan courseId wajib")
			return
		}
		if _, err := uuid.Parse(userID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "userId tidak valid")
			return
		}
		if _, err := uuid.Parse(courseID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "courseId tidak valid")
			return
		}
		if _, err := deps.UserRepo.FindByID(r.Context(), userID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "user_not_found", "pengguna tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if _, err := deps.CourseRepo.GetByID(r.Context(), courseID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "course_not_found", "kelas tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		if _, err := deps.EnrollmentRepo.GetByUserAndCourse(r.Context(), userID, courseID); err == nil {
			writeError(w, http.StatusConflict, "already_enrolled", "siswa sudah terdaftar di kelas ini")
			return
		} else if !errors.Is(err, pgx.ErrNoRows) {
			writeInternalError(w, r, err)
			return
		}
		enrolledAt := time.Now()
		if req.EnrolledAt != nil && strings.TrimSpace(*req.EnrolledAt) != "" {
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(*req.EnrolledAt))
			if err != nil {
				writeError(w, http.StatusBadRequest, "bad_request", "enrolledAt harus RFC3339")
				return
			}
			enrolledAt = t
		}
		created, err := deps.EnrollmentRepo.Create(r.Context(), domain.CourseEnrollment{
			UserID:     userID,
			CourseID:   courseID,
			Status:     domain.EnrollmentStatusEnrolled,
			EnrolledAt: enrolledAt,
		})
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enrollmentId": created.ID,
			"userId":       created.UserID,
			"courseId":     created.CourseID,
			"status":       created.Status,
			"enrolledAt":   created.EnrolledAt.Format(time.RFC3339),
		})
	}
}

// AdminPatchEnrollment PATCH /api/v1/admin/enrollments/{enrollmentId} — ubah tanggal enrolled_at.
func AdminPatchEnrollment(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enrollmentID := strings.TrimSpace(chi.URLParam(r, "enrollmentId"))
		if _, err := uuid.Parse(enrollmentID); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "enrollmentId tidak valid")
			return
		}
		var req dto.AdminPatchEnrollmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "body tidak valid")
			return
		}
		if strings.TrimSpace(req.EnrolledAt) == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "enrolledAt wajib (RFC3339)")
			return
		}
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(req.EnrolledAt))
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "enrolledAt harus RFC3339")
			return
		}
		if err := deps.EnrollmentRepo.UpdateEnrolledAt(r.Context(), enrollmentID, t); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeError(w, http.StatusNotFound, "not_found", "enrollment tidak ditemukan")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "Tanggal enrollment diperbarui"})
	}
}
