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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tryoutToDTO(t))
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
		attempt, err := deps.AttemptService.Start(r.Context(), userID, tryoutID)
		if err != nil {
			if err == service.ErrAlreadySubmitted {
				http.Error(w, "already submitted", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t, _ := deps.TryoutService.GetByID(r.Context(), tryoutID)
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
		OpensAt:         t.OpensAt,
		ClosesAt:        t.ClosesAt,
		MaxParticipants: t.MaxParticipants,
		Status:          t.Status,
	}
}
