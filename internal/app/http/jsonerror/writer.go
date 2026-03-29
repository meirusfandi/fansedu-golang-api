// Package jsonerror writes a single JSON error shape for all API responses:
//
//	{ "error": { "code": "STABLE_CODE", "message": "Safe user-facing text" } }
package jsonerror

import (
	"encoding/json"
	"net/http"
	"strings"
)

type payload struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NormalizeCode turns any input into UPPER_SNAKE stable codes for clients.
func NormalizeCode(code string) string {
	s := strings.TrimSpace(strings.ReplaceAll(code, "-", "_"))
	if s == "" {
		return "ERROR"
	}
	return strings.ToUpper(s)
}

// Write sends a JSON error body. message must be safe for end users (no SQL, paths, stack).
func Write(w http.ResponseWriter, status int, code, message string) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = http.StatusText(status)
	}
	if msg == "" {
		msg = "Permintaan tidak dapat diproses."
	}
	p := payload{}
	p.Error.Code = NormalizeCode(code)
	p.Error.Message = msg
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}
