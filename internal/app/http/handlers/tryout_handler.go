package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
	"github.com/meirusfandi/fansedu-golang-api/internal/cache"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func TryoutListOpen(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := deps.TryoutService.ListOpen(r.Context())
		if err != nil {
			writeInternalError(w, r, err)
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

// TryoutList: GET /api/v1/tryouts?status=open — daftar tryout berstatus open dengan closes_at belum lewat.
func TryoutList(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		if status == "" || status == domain.TryoutStatusOpen || status == "open" {
			list, err := deps.TryoutService.ListOpen(r.Context())
			if err != nil {
				writeInternalError(w, r, err)
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
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}

		// Subject guard (student only)
		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}

		if !canRegisterForTryout(t, time.Now()) {
			writeError(w, http.StatusForbidden, "REGISTRATION_CLOSED", "Pendaftaran tryout tidak tersedia.")
			return
		}

		already, err := deps.TryoutRegistrationRepo.IsRegistered(r.Context(), userID, tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}

		if err := deps.TryoutService.Register(r.Context(), userID, tryoutID); err != nil {
			writeInternalError(w, r, err)
			return
		}

		registeredAt, ok, err := deps.TryoutRegistrationRepo.GetRegisteredAt(r.Context(), userID, tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		if !ok {
			log.Printf("tryout register: registered_at missing tryout=%s user=%s", tryoutID, userID)
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server.")
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
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}

		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}

		attempt, ok := startTryoutExamForUser(w, r, deps, tryoutID, userID)
		if !ok {
			return
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

		ctx := r.Context()
		t, err := deps.TryoutService.GetByID(ctx, tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		if role, _ := middleware.GetRole(ctx); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, uerr := deps.UserRepo.FindByID(ctx, userID)
				if uerr != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}

		isRegistered, err := deps.TryoutRegistrationRepo.IsRegistered(ctx, userID, tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}

		attempts, err := deps.AttemptService.ListByUser(ctx, userID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}

		count := 0
		hasAttempted := false
		var lastAttemptID *string
		var lastAttemptTime time.Time
		var latestForTryout *domain.Attempt
		for _, a := range attempts {
			if a.TryoutSessionID != tryoutID {
				continue
			}
			if latestForTryout == nil {
				a2 := a
				latestForTryout = &a2
			}
			count++
			hasAttempted = true
			if lastAttemptID == nil || a.StartedAt.After(lastAttemptTime) {
				t0 := a.StartedAt
				lastAttemptTime = t0
				id := a.ID
				lastAttemptID = &id
			}
		}

		canRetake := hasAttempted
		now := time.Now()
		canRegister := !isRegistered && canRegisterForTryout(t, now)
		canStartExam := false
		startReason := ""
		if !isRegistered {
			startReason = "NOT_REGISTERED"
		} else if latestForTryout != nil {
			switch latestForTryout.Status {
			case domain.AttemptStatusInProgress:
				canStartExam = true
			case domain.AttemptStatusSubmitted:
				startReason = "ALREADY_SUBMITTED"
			default:
				if canStartTryoutExam(t, now) {
					canStartExam = true
				} else {
					startReason, _ = tryoutStartBlockReason(t, now)
				}
			}
		} else {
			if canStartTryoutExam(t, now) {
				canStartExam = true
			} else {
				startReason, _ = tryoutStartBlockReason(t, now)
			}
		}

		resp := dto.StudentTryoutStatusResponse{
			IsRegistered:        isRegistered,
			HasAttempted:        hasAttempted,
			CanRetake:           canRetake,
			AttemptCount:        count,
			LastAttemptID:       lastAttemptID,
			OpensAt:             t.OpensAt.UTC().Format(time.RFC3339),
			ClosesAt:            t.ClosesAt.UTC().Format(time.RFC3339),
			TryoutStatus:        t.Status,
			CanRegister:         canRegister,
			CanStartExam:        canStartExam,
			StartDisabledReason: startReason,
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}

		attempts, err := deps.AttemptService.ListByUser(r.Context(), userID)
		if err != nil {
			writeInternalError(w, r, err)
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		var subjectID *string
		if u, err := deps.UserRepo.FindByID(r.Context(), userID); err == nil {
			subjectID = u.SubjectID
		}
		list, err := deps.TryoutService.ListForStudent(r.Context(), subjectID)
		if err != nil {
			writeInternalError(w, r, err)
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

// StudentTryoutListOpen: tryout status open untuk bidang siswa; closes_at belum lewat. Requires Auth.
func StudentTryoutListOpen(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		var subjectID *string
		if u, err := deps.UserRepo.FindByID(r.Context(), userID); err == nil {
			subjectID = u.SubjectID
		}
		list, err := deps.TryoutService.ListOpenForStudent(r.Context(), subjectID)
		if err != nil {
			writeInternalError(w, r, err)
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
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		// Siswa hanya boleh melihat tryout yang sesuai bidang-nya (atau tryout umum subject_id = nil)
		if userID, ok := middleware.GetUserID(r.Context()); ok && userID != "" {
			if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
				if t.SubjectID != nil && *t.SubjectID != "" {
					u, err := deps.UserRepo.FindByID(r.Context(), userID)
					if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
						writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		if tryoutID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
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
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		if tryoutID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}
		if !canRegisterForTryout(t, time.Now()) {
			writeError(w, http.StatusForbidden, "REGISTRATION_CLOSED", "Pendaftaran tryout tidak tersedia.")
			return
		}
		if err := deps.TryoutService.Register(r.Context(), userID, tryoutID); err != nil {
			writeInternalError(w, r, err)
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
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		if _, err := deps.TryoutService.GetByID(r.Context(), tryoutID); err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		list, err := deps.TryoutService.GetLeaderboard(r.Context(), tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		if list == nil {
			list = []domain.LeaderboardEntry{}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(list)
	}
}

// TryoutLeaderboardTop GET /api/v1/tryouts/{tryoutId}/leaderboard/top — top N dari Redis sorted set (cepat).
// Query: limit (default 10, max 100).
func TryoutLeaderboardTop(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		if tryoutID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		if _, err := deps.TryoutService.GetByID(r.Context(), tryoutID); err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		limit := 10
		if v := r.URL.Query().Get("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
				limit = n
			}
		}
		ctx := r.Context()
		fetch := limit * 15
		if fetch < 60 {
			fetch = 60
		}
		if fetch > 500 {
			fetch = 500
		}
		zs, err := cache.LeaderboardTop(ctx, deps.Redis, tryoutID, fetch)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := make([]dto.LeaderboardTopRow, 0, limit)
		for _, z := range zs {
			if len(out) >= limit {
				break
			}
			uid, _ := z.Member.(string)
			reg, err := deps.TryoutRegistrationRepo.IsRegistered(ctx, uid, tryoutID)
			if err != nil {
				writeInternalError(w, r, err)
				return
			}
			if !reg {
				continue
			}
			name := uid
			if u, err := deps.UserRepo.FindByID(ctx, uid); err == nil {
				name = u.Name
			}
			out = append(out, dto.LeaderboardTopRow{
				Rank:     len(out) + 1,
				UserID:   uid,
				UserName: name,
				Score:    z.Score,
			})
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// TryoutLeaderboardMyRank GET /api/v1/tryouts/{tryoutId}/leaderboard/rank — peringkat user dari ZSET (Bearer).
func TryoutLeaderboardMyRank(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		ctx := r.Context()
		if _, err := deps.TryoutService.GetByID(ctx, tryoutID); err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		reg, err := deps.TryoutRegistrationRepo.IsRegistered(ctx, userID, tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		if !reg {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(dto.LeaderboardRankResponse{InLeaderboard: false})
			return
		}
		rank0, score, ok, err := cache.LeaderboardUserRankScore(ctx, deps.Redis, tryoutID, userID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		resp := dto.LeaderboardRankResponse{InLeaderboard: ok}
		if ok {
			resp.Rank = int(rank0) + 1
			resp.Score = score
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TryoutStart(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		userID, _ := middleware.GetUserID(r.Context())
		if userID == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Autentikasi diperlukan.")
			return
		}
		t, err := deps.TryoutService.GetByID(r.Context(), tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		if role, _ := middleware.GetRole(r.Context()); domain.IsStudentRoleCode(role) {
			if t.SubjectID != nil && *t.SubjectID != "" {
				u, err := deps.UserRepo.FindByID(r.Context(), userID)
				if err != nil || u.SubjectID == nil || *u.SubjectID != *t.SubjectID {
					writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
					return
				}
			}
		}
		attempt, ok := startTryoutExamForUser(w, r, deps, tryoutID, userID)
		if !ok {
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
	gm := t.GradingMode
	if gm == "" {
		gm = domain.TryoutGradingModeAuto
	}
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
		GradingMode:     gm,
	}
}
