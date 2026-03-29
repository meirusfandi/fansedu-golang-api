package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/dto"
)

// TrainerGuruTryoutPaperGet — GET .../guru|trainer/tryouts/:tryoutId/paper
func TrainerGuruTryoutPaperGet(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tryoutID := chi.URLParam(r, "tryoutId")
		if tryoutID == "" {
			writeError(w, http.StatusBadRequest, "BAD_REQUEST", "ID tryout wajib diisi.")
			return
		}
		ctx := r.Context()
		t, err := deps.TryoutService.GetByID(ctx, tryoutID)
		if err != nil {
			writeError(w, http.StatusNotFound, "TRYOUT_NOT_FOUND", "Tryout tidak ditemukan.")
			return
		}
		qs, err := deps.QuestionRepo.ListByTryoutSessionID(ctx, tryoutID)
		if err != nil {
			writeInternalError(w, r, err)
			return
		}
		out := dto.GuruTryoutPaperResponse{
			Title:           t.Title,
			DurationMinutes: t.DurationMinutes,
			OpensAt:         t.OpensAt,
			ClosesAt:        t.ClosesAt,
			Questions:       make([]dto.QuestionResponse, len(qs)),
		}
		for i := range qs {
			out.Questions[i] = questionToDTO(qs[i])
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// TrainerGuruTryoutPaperPut — PUT bulk lembar soal; belum diimplementasikan (gunakan admin).
func TrainerGuruTryoutPaperPut(deps *Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = deps
		writeError(w, http.StatusNotImplemented, "FEATURE_NOT_IMPLEMENTED", "Pembaruan lembar soal melalui API ini belum tersedia.")
	}
}
