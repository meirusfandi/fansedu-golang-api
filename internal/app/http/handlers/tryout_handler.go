package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

func TryoutListOpen(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.TryoutService.ListOpen(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.TryoutResponse, len(list))
		for i := range list {
			out[i] = tryoutToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// StudentTryoutList returns all tryouts/events for the student's subject (bidang), excluding draft.
// Status open/closed and opens_at/closes_at are included; frontend can separate "dibuka" vs "ditutup". Requires Auth.
func StudentTryoutList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var subjectID *string
		if u, err := deps.UserRepo.FindByID(r.Context(), userID); err == nil {
			subjectID = u.SubjectID
		}
		list, err := deps.TryoutService.ListForStudent(r.Context(), subjectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.TryoutResponse, len(list))
		for i := range list {
			out[i] = tryoutToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// StudentTryoutListOpen returns only currently open tryouts (by time window) for the student's subject. For dashboard widget. Requires Auth.
func StudentTryoutListOpen(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		var subjectID *string
		if u, err := deps.UserRepo.FindByID(r.Context(), userID); err == nil {
			subjectID = u.SubjectID
		}
		list, err := deps.TryoutService.ListOpenForStudent(r.Context(), subjectID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out := make([]dto.TryoutResponse, len(list))
		for i := range list {
			out[i] = tryoutToDTO(list[i])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func TryoutGetByID(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "tryoutId")
		if id == "" {
			http.Error(w, "tryout id required", http.StatusBadRequest)
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, "tryout not found", http.StatusNotFound)
			return
		}
		// Siswa hanya boleh melihat tryout yang sesuai bidang-nya (atau tryout umum subject_id = nil)
		if userID, ok := middleware.GetUserID(r.Context()); ok && userID != "" {
			if role, _ := middleware.GetRole(r.Context()); role == "student" {
				if t.SubjectID != nil && *t.SubjectID != "" {
					u, err := deps.UserRepo.FindByID(r.Context(), userID)
					if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
						http.Error(w, "tryout not found", http.StatusNotFound)
						return
					}
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tryoutToDTO(t))
	}
}

// TryoutRegister mendaftarkan siswa ke tryout (masuk ke leaderboard). Auth required.
func TryoutRegister(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if tryoutID == "" {
			http.Error(w, "tryout id required", http.StatusBadRequest)
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, "tryout not found", http.StatusNotFound)
			return
		}
		if role, _ := middleware.GetRole(r.Context()); role == "student" {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					http.Error(w, "tryout not found", http.StatusNotFound)
					return
				}
			}
		}
		if err := deps.TryoutService.Register(r.Context(), userID, tryoutID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "registered"})
	}
}

// TryoutLeaderboard mengembalikan leaderboard tryout: urutan nama (belum mengerjakan), lalu nilai tertinggi, waktu tercepat, nama.
func TryoutLeaderboard(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		if tryoutID == "" {
			http.Error(w, "tryout id required", http.StatusBadRequest)
			return
		}
		if _, err := deps.TryoutService.GetByID(r.Context(), tryoutID); err != nil {
			http.Error(w, "tryout not found", http.StatusNotFound)
			return
		}
		list, err := deps.TryoutService.GetLeaderboard(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}
}

func TryoutStart(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, "tryout not found", http.StatusNotFound)
			return
		}
		if role, _ := middleware.GetRole(r.Context()); role == "student" {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					http.Error(w, "tryout not found", http.StatusNotFound)
					return
				}
			}
		}
		// Auto-register ke tryout agar masuk leaderboard bila belum terdaftar
		_ = deps.TryoutRegistrationRepo.Register(r.Context(), userID, tryoutID)
		attempt, err := deps.AttemptService.Start(r.Context(), userID, tryoutID)
		if err != nil {
			if err == service.ErrAlreadySubmitted {
				http.Error(w, "already submitted", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expiresAt := attempt.StartedAt.Add(time.Duration(t.DurationMinutes) * time.Minute)
		timeLeft := int(time.Until(expiresAt).Seconds())
		if timeLeft < 0 {
			timeLeft = 0
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.TryoutStartResponse{
			AttemptID:       attempt.ID,
			ExpiresAt:       expiresAt,
			TimeLeftSeconds: timeLeft,
		})
	}
}

func tryoutToDTO(t domain.TryoutSession) dto.TryoutResponse {
	return dto.TryoutResponse{
		ID:              t.ID,
		Title:           t.Title,
		ShortTitle:      t.ShortTitle,
		Description:     t.Description,
		DurationMinutes: t.DurationMinutes,
		QuestionsCount:  t.QuestionsCount,
		Level:           t.Level,
		SubjectID:       t.SubjectID,
		OpensAt:         t.OpensAt,
		ClosesAt:        t.ClosesAt,
		MaxParticipants: t.MaxParticipants,
		Status:          t.Status,
	}
}
