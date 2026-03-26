package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

// WriteJSONError mengirim {"error","message"} agar konsisten dengan handler (FE bisa parse selalu).
func WriteJSONError(w http.ResponseWriter, status int, code, message string) {
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
