package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func decodeAttemptAnswerReviewPatch(r *http.Request) (service.AttemptAnswerReviewPatch, error) {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return service.AttemptAnswerReviewPatch{}, err
	}
	var patch service.AttemptAnswerReviewPatch
	if b, ok := raw["reviewerComment"]; ok {
		patch.HasReviewerComment = true
		if string(b) == "null" {
			patch.ReviewerComment = nil
		} else {
			var s string
			if err := json.Unmarshal(b, &s); err != nil {
				return patch, err
			}
			patch.ReviewerComment = &s
		}
	}
	if b, ok := raw["manualScore"]; ok {
		patch.HasManualScore = true
		if string(b) == "null" {
			patch.ManualScore = nil
		} else {
			var f float64
			if err := json.Unmarshal(b, &f); err != nil {
				return patch, err
			}
			patch.ManualScore = &f
		}
	}
	return patch, nil
}

// AdminGetAttemptReview GET /api/v1/admin/tryouts/{tryoutId}/attempts/{attemptId}/review
func AdminGetAttemptReview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		out, err := deps.AdminService.GetAttemptReview(r.Context(), tryoutID, attemptID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// AdminPutAttemptAnswerReview PUT /api/v1/admin/tryouts/{tryoutId}/attempts/{attemptId}/answers/{questionId}/review
func AdminPutAttemptAnswerReview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		questionID := chi.URLParam(r, "questionId")
		reviewerID, _ := middleware.GetUserID(r.Context())
		if reviewerID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		patch, err := decodeAttemptAnswerReviewPatch(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body permintaan tidak valid.")
			return
		}
		studentUID, newScore, err := deps.AdminService.PutAttemptAnswerReview(r.Context(), tryoutID, attemptID, questionID, reviewerID, patch)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if errors.Is(err, service.ErrAttemptReviewNoFields) {
				writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Kirim reviewerComment dan/atau manualScore.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		syncLeaderboardAfterReview(r.Context(), deps, tryoutID, studentUID, newScore)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":         true,
			"attemptId":  attemptID,
			"questionId": questionID,
			"score":      newScore,
		})
	}
}

func tryoutTrainerSubjectGuard(ctx context.Context, deps *Deps, userID, tryoutID string) bool {
	t, err := deps.TryoutService.GetByID(ctx, tryoutID)
	if err != nil {
		return false
	}
	u, err := deps.UserRepo.FindByID(ctx, userID)
	if err != nil {
		return false
	}
	if t.SubjectID != nil && *t.SubjectID != "" {
		if u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
			return false
		}
	}
	return true
}

// TrainerGetAttemptReview GET /api/v1/trainer|guru/tryouts/{tryoutId}/attempts/{attemptId}/review
func TrainerGetAttemptReview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		if !tryoutTrainerSubjectGuard(r.Context(), deps, userID, tryoutID) {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		out, err := deps.AdminService.GetAttemptReview(r.Context(), tryoutID, attemptID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// TrainerPutAttemptAnswerReview PUT /api/v1/trainer|guru/tryouts/.../answers/{questionId}/review
func TrainerPutAttemptAnswerReview(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		questionID := chi.URLParam(r, "questionId")
		if !tryoutTrainerSubjectGuard(r.Context(), deps, userID, tryoutID) {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		patch, err := decodeAttemptAnswerReviewPatch(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body permintaan tidak valid.")
			return
		}
		studentUID, newScore, err := deps.AdminService.PutAttemptAnswerReview(r.Context(), tryoutID, attemptID, questionID, userID, patch)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan.")
				return
			}
			if errors.Is(err, service.ErrAttemptReviewNoFields) {
				writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Kirim reviewerComment dan/atau manualScore.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		syncLeaderboardAfterReview(r.Context(), deps, tryoutID, studentUID, newScore)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":         true,
			"attemptId":  attemptID,
			"questionId": questionID,
			"score":      newScore,
		})
	}
}

func syncLeaderboardAfterReview(ctx context.Context, deps *Deps, tryoutID, studentUserID string, newScore float64) {
	if deps.Redis == nil {
		return
	}
	reg, err := deps.TryoutRegistrationRepo.IsRegistered(ctx, studentUserID, tryoutID)
	if err != nil || !reg {
		return
	}
	_ = cache.LeaderboardZAdd(ctx, deps.Redis, tryoutID, studentUserID, newScore)
}

// AdminPostAttemptAutoGrade POST /api/v1/admin/tryouts/{tryoutId}/attempts/{attemptId}/auto-grade
// Hapus semua manual_score, lalu hitung ulang dari kunci soal. Body opsional: { "clearReviewerComments": true }.
func AdminPostAttemptAutoGrade(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		var opts service.AutoGradeAttemptOpts
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&opts); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body JSON tidak valid.")
			return
		}
		studentUID, newScore, err := deps.AdminService.AutoGradeAttempt(r.Context(), tryoutID, attemptID, opts)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			writeInternalError(w, r, err)
			return
		}
		syncLeaderboardAfterReview(r.Context(), deps, tryoutID, studentUID, newScore)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"attemptId": attemptID,
			"score":     newScore,
		})
	}
}

// TrainerPostAttemptAutoGrade POST /api/v1/trainer|guru/tryouts/.../attempts/.../auto-grade
func TrainerPostAttemptAutoGrade(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		tryoutID := chi.URLParam(r, "tryoutId")
		attemptID := chi.URLParam(r, "attemptId")
		if !tryoutTrainerSubjectGuard(r.Context(), deps, userID, tryoutID) {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		var opts service.AutoGradeAttemptOpts
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&opts); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Body JSON tidak valid.")
			return
		}
		studentUID, newScore, err := deps.AdminService.AutoGradeAttempt(r.Context(), tryoutID, attemptID, opts)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				writeError(w, http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan.")
				return
			}
			writeInternalError(w, r, err)
			return
		}
		syncLeaderboardAfterReview(r.Context(), deps, tryoutID, studentUID, newScore)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"attemptId": attemptID,
			"score":     newScore,
		})
	}
}
