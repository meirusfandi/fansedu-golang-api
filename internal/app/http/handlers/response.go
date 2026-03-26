package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// writeError sends a JSON error response matching spec: { "error": "code", "message": "..." }
func writeError(w http.ResponseWriter, status int, code, message string) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = http.StatusText(status)
	}
	if msg == "" {
		msg = "Request failed"
	}
	if strings.TrimSpace(code) == "" {
		code = "error"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{Error: code, Message: msg})
}

// writeErrorFromUserRepoUpdate maps user UPDATE failures (e.g. no matching row) to JSON errors.
func writeErrorFromUserRepoUpdate(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not_found", "user not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "server_error", err.Error())
}
