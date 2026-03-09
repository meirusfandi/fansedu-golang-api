package handlers

import (
	"encoding/json"
	"net/http"
)

// writeError sends a JSON error response matching spec: { "error": "code", "message": "..." }
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}{Error: code, Message: message})
}
