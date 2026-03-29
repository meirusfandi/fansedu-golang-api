package middleware

import (
	"net/http"
	"strings"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/jsonerror"
)

// WriteJSONError sends { "error": { "code", "message" } } with stable uppercase codes.
func WriteJSONError(w http.ResponseWriter, status int, code, message string) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = http.StatusText(status)
	}
	if msg == "" {
		msg = "Permintaan tidak dapat diproses."
	}
	if strings.TrimSpace(code) == "" {
		code = "ERROR"
	}
	jsonerror.Write(w, status, code, msg)
}
