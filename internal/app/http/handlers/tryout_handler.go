package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
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

// TryoutList supports query param: GET /api/v1/tryouts?status=open
// Currently we only support status=open (public).
func TryoutList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		if status == "" || status == domain.TryoutStatusOpen || status == "open" {
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
			return
		}
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid status; only status=open is supported")
	}
}

// StudentTryoutRegister registers student to a tryout from student tryout detail.
// POST /api/v1/student/tryouts/:tryoutId/register
func StudentTryoutRegister(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" || tryoutID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "userId/tryoutId required")
			return
		}

		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			http.Error(w, "tryout not found", http.StatusNotFound)
			return
		}

		// Subject guard (student only)
		if role, _ := middleware.GetRole(r.Context()); role == "student" {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					http.Error(w, "tryout not found", http.StatusNotFound)
					return
				}
			}
		}

		already, err := deps.TryoutRegistrationRepo.IsRegistered(r.Context(), userID, tryoutID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		if err := deps.TryoutService.Register(r.Context(), userID, tryoutID); err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		registeredAt, ok, err := deps.TryoutRegistrationRepo.GetRegisteredAt(r.Context(), userID, tryoutID)
		if err != nil || !ok {
			writeError(w, http.StatusInternalServerError, "server_error", "failed to fetch registered_at")
			return
		}

		if already {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"registered":    true,
			"tryoutId":      tryoutID,
			"registeredAt": registeredAt.UTC().Format(time.RFC3339),
		})
	}
}

// StudentTryoutStart starts an exam/attempt for the student.
// POST /api/v1/student/tryouts/:tryoutId/start
func StudentTryoutStart(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" || tryoutID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "userId/tryoutId required")
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

		attempt, err := deps.AttemptService.Start(r.Context(), userID, tryoutID)
		if err != nil {
			if err == service.ErrAlreadySubmitted {
				// If already started/submitted, return the latest existing attempt.
				attempts, _ := deps.AttemptService.ListByUser(r.Context(), userID)
				found := domain.Attempt{}
				ok := false
				for _, a := range attempts {
					if a.TryoutSessionID == tryoutID {
						found = a
						ok = true
						break
					}
				}
				if !ok {
					writeError(w, http.StatusConflict, "already_submitted", "attempt already submitted")
					return
				}
				attempt = found
			} else {
				writeError(w, http.StatusInternalServerError, "server_error", err.Error())
				return
			}
		}

		base := os.Getenv("EXAM_BASE_URL")
		if base == "" {
			base = "https://exam.fansedu.id"
		}
		base = strings.TrimRight(base, "/")
		examURL := base + "/attempt/" + attempt.ID

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"attemptId": attempt.ID,
			"examUrl":   examURL,
			"startedAt": attempt.StartedAt.UTC().Format(time.RFC3339),
		})
	}
}

// GET /api/v1/student/tryouts/:tryoutId/status
// Returns registration/attempt state to drive button logic.
func StudentTryoutStatus(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" || tryoutID == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "userId/tryoutId required")
			return
		}

		isRegistered, err := deps.TryoutRegistrationRepo.IsRegistered(r.Context(), userID, tryoutID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		attempts, err := deps.AttemptService.ListByUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		count := 0
		hasAttempted := false
		var lastAttemptID *string
		var lastAttemptTime time.Time
		for _, a := range attempts {
			if a.TryoutSessionID != tryoutID {
				continue
			}
			count++
			hasAttempted = true
			if lastAttemptID == nil || a.StartedAt.After(lastAttemptTime) {
				t := a.StartedAt
				lastAttemptTime = t
				id := a.ID
				lastAttemptID = &id
			}
		}

		canRetake := hasAttempted
		resp := dto.StudentTryoutStatusResponse{
			IsRegistered: isRegistered,
			HasAttempted: hasAttempted,
			CanRetake:    canRetake,
			AttemptCount: count,
			LastAttemptID: lastAttemptID,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// GET /api/v1/student/tryouts/history
// Returns most recent submitted attempts per tryout with score improvement.
func StudentTryoutHistory(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok || userID == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "not authenticated")
			return
		}

		attempts, err := deps.AttemptService.ListByUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "server_error", err.Error())
			return
		}

		// group submitted attempts by tryoutId
		type scoreAttempt struct {
			attemptID string
			score     float64
			submittedAt time.Time
		}
		group := map[string][]scoreAttempt{}
		for _, a := range attempts {
			if a.Status != domain.AttemptStatusSubmitted || a.TryoutSessionID == "" || a.SubmittedAt == nil {
				continue
			}
			score := 0.0
			if a.Score != nil {
				score = *a.Score
			}
			sa := scoreAttempt{
				attemptID:   a.ID,
				score:       score,
				submittedAt: *a.SubmittedAt,
			}
			group[a.TryoutSessionID] = append(group[a.TryoutSessionID], sa)
		}

		// pick latest per tryout
		type historyRow struct {
			item dto.StudentTryoutHistoryItem
			at   time.Time
		}
		rows := make([]historyRow, 0, len(group))
		tryoutTitleCache := map[string]string{}
		for tryoutID, list := range group {
			// sort desc by submittedAt
			for i := 0; i < len(list); i++ {
				for j := i + 1; j < len(list); j++ {
					if list[j].submittedAt.After(list[i].submittedAt) {
						list[i], list[j] = list[j], list[i]
					}
				}
			}
			latest := list[0]
			prevScore := 0.0
			if len(list) > 1 {
				prevScore = list[1].score
			}
			improvement := latest.score - prevScore

			title, ok := tryoutTitleCache[tryoutID]
			if !ok {
				t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
				if err == nil {
					title = t.Title
				} else {
					title = ""
				}
				tryoutTitleCache[tryoutID] = title
			}

			item := dto.StudentTryoutHistoryItem{
				TryoutID: tryoutID,
				TryoutTitle: title,
				AttemptID: latest.attemptID,
				Score: latest.score,
				SubmittedAt: latest.submittedAt.UTC().Format(time.RFC3339),
				ImprovementFromPrevious: improvement,
			}
			rows = append(rows, historyRow{item: item, at: latest.submittedAt})
		}

		// sort rows desc by last submittedAt
		for i := 0; i < len(rows); i++ {
			for j := i + 1; j < len(rows); j++ {
				if rows[j].at.After(rows[i].at) {
					rows[i], rows[j] = rows[j], rows[i]
				}
			}
		}

		// limit
		if len(rows) > 20 {
			rows = rows[:20]
		}

		out := make([]dto.StudentTryoutHistoryItem, 0, len(rows))
		for _, r := range rows {
			out = append(out, r.item)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(dto.StudentTryoutHistoryResponse{Data: out})
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

// StudentTryoutGetByID detail tryout untuk siswa. Auth wajib; hanya tryout yang sesuai subject siswa.
// Dipanggil dari halaman student/tryouts/:id (frontend route student/tryouts/id).
func StudentTryoutGetByID(deps *Deps) http.HandlerFunc {
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
		if list == nil {
			list = []domain.LeaderboardEntry{}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
