package handlers

import (
	"net/http"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
	"github.com/meirusfandi/fansedu-golang-api/internal/service"
)

// canRegisterForTryout: siswa boleh "daftar" — tryout open dan belum lewat closesAt.
func canRegisterForTryout(t domain.TryoutSession, now time.Time) bool {
	if t.Status != domain.TryoutStatusOpen {
		return false
	}
	return !now.UTC().After(t.ClosesAt.UTC())
}

// canStartTryoutExam: boleh memulai attempt baru — status open dan sekarang di [opensAt, closesAt].
func canStartTryoutExam(t domain.TryoutSession, now time.Time) bool {
	if t.Status != domain.TryoutStatusOpen {
		return false
	}
	u := now.UTC()
	return !u.Before(t.OpensAt.UTC()) && !u.After(t.ClosesAt.UTC())
}

func tryoutStartBlockReason(t domain.TryoutSession, now time.Time) (code, message string) {
	if t.Status != domain.TryoutStatusOpen {
		return "TRYOUT_NOT_OPEN", "Tryout tidak dibuka untuk pengerjaan."
	}
	u := now.UTC()
	if u.Before(t.OpensAt.UTC()) {
		return "BEFORE_OPENS_AT", "Waktu mulai tryout belum dimulai."
	}
	if u.After(t.ClosesAt.UTC()) {
		return "AFTER_CLOSES_AT", "Batas waktu tryout sudah berakhir."
	}
	return "FORBIDDEN", "Tidak dapat memulai ujian."
}

// startTryoutExamForUser: wajib terdaftar; resume in_progress selalu boleh; attempt baru hanya dalam jendela waktu.
func startTryoutExamForUser(w http.ResponseWriter, r *http.Request, deps *Deps, tryoutID, userID string) (domain.Attempt, bool) {
	ctx := r.Context()
	reg, err := deps.TryoutRegistrationRepo.IsRegistered(ctx, userID, tryoutID)
	if err != nil {
		writeInternalError(w, r, err)
		return domain.Attempt{}, false
	}
	if !reg {
		writeError(w, http.StatusForbidden, "NOT_REGISTERED", "Daftar tryout terlebih dahulu agar bisa mulai ujian.")
		return domain.Attempt{}, false
	}
	t, err := deps.TryoutService.GetByID(ctx, tryoutID)
	if err != nil {
		writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
		return domain.Attempt{}, false
	}
	attempts, err := deps.AttemptService.ListByUser(ctx, userID)
	if err != nil {
		writeInternalError(w, r, err)
		return domain.Attempt{}, false
	}
	var latest *domain.Attempt
	for i := range attempts {
		if attempts[i].TryoutSessionID == tryoutID {
			latest = &attempts[i]
			break
		}
	}
	now := time.Now()
	if latest != nil {
		if latest.Status == domain.AttemptStatusInProgress {
			return *latest, true
		}
		if latest.Status == domain.AttemptStatusSubmitted {
			writeError(w, http.StatusConflict, "ALREADY_SUBMITTED", "Tryout sudah diselesaikan.")
			return domain.Attempt{}, false
		}
	}
	if !canStartTryoutExam(t, now) {
		code, msg := tryoutStartBlockReason(t, now)
		writeError(w, http.StatusForbidden, code, msg)
		return domain.Attempt{}, false
	}
	attempt, err := deps.AttemptService.Start(ctx, userID, tryoutID)
	if err != nil {
		if err == service.ErrAlreadySubmitted {
			writeError(w, http.StatusConflict, "ALREADY_SUBMITTED", "Tryout sudah diselesaikan.")
			return domain.Attempt{}, false
		}
		writeInternalError(w, r, err)
		return domain.Attempt{}, false
	}
	return attempt, true
}
