package handlers

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/jsonerror"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

// writeError sends { "error": { "code", "message" } }. Use only safe, non-technical messages for clients.
func writeError(w http.ResponseWriter, status int, code, message string) {
	jsonerror.Write(w, status, code, message)
}

// writeInternalError logs the real error and returns a generic 500 JSON response (unless EXPOSE_INTERNAL_ERRORS=true).
func writeInternalError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server.")
		return
	}
	rid, _ := middleware.GetRequestID(r.Context())
	log.Printf("api internal error path=%s rid=%s err=%v", r.URL.Path, rid, err)
	msg := "Terjadi kesalahan pada server."
	if exposeInternalErrors() {
		msg = err.Error()
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}

func writeInternalErrorNoReq(w http.ResponseWriter, err error) {
	if err == nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server.")
		return
	}
	log.Printf("api internal error (no req): %v", err)
	msg := "Terjadi kesalahan pada server."
	if exposeInternalErrors() {
		msg = err.Error()
	}
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}

func exposeInternalErrors() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("EXPOSE_INTERNAL_ERRORS")), "true")
}

// writeErrorFromUserRepoUpdate maps user UPDATE failures (e.g. no matching row) to JSON errors.
func writeErrorFromUserRepoUpdate(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Pengguna tidak ditemukan.")
		return
	}
	writeInternalError(w, r, err)
}
