package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

// ApplicationErrorLogInserter disuntikkan ke middleware (repo); nil = no-op.
type ApplicationErrorLogInserter interface {
	Insert(ctx context.Context, e domain.ApplicationErrorLog) error
}

const maxErrorBodyCapture = 8192

type captureResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	buf         *bytes.Buffer
}

func (w *captureResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
		w.status = http.StatusOK
	}
	if w.status >= 400 && w.buf != nil && w.buf.Len() < maxErrorBodyCapture {
		if space := maxErrorBodyCapture - w.buf.Len(); space > 0 {
			if len(b) > space {
				_, _ = w.buf.Write(b[:space])
			} else {
				_, _ = w.buf.Write(b)
			}
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *captureResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// ErrorResponseLogger mencatat respons HTTP status >= 400 ke application_error_logs (async).
// Berlaku untuk semua role: user_id / user_role diambil dari context bila sudah di-set middleware Auth.
func ErrorResponseLogger(inserter ApplicationErrorLogInserter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if inserter == nil || shouldSkipErrorLogPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			buf := &bytes.Buffer{}
			cw := &captureResponseWriter{ResponseWriter: w, buf: buf}
			next.ServeHTTP(cw, r)
			st := cw.status
			if st == 0 {
				st = http.StatusOK
			}
			if st < 400 {
				return
			}
			errCode, msg := parseErrorJSON(buf.Bytes())
			etype := inferApplicationErrorType(st, errCode)
			if msg == "" {
				msg = http.StatusText(st)
				if msg == "" {
					msg = "HTTP error"
				}
			}
			userID, _ := GetUserID(r.Context())
			role, _ := GetRole(r.Context())
			reqID, _ := GetRequestID(r.Context())
			var uid *string
			if userID != "" {
				uid = &userID
			}
			var urole *string
			if role != "" {
				urole = &role
			}
			var rid *string
			if reqID != "" {
				rid = &reqID
			}
			ip := strings.TrimSpace(r.RemoteAddr)
			var pip *string
			if ip != "" {
				pip = &ip
			}
			ua := strings.TrimSpace(r.UserAgent())
			var pu *string
			if ua != "" {
				pu = &ua
			}
			qs := r.URL.RawQuery
			var pqs *string
			if qs != "" {
				pqs = &qs
			}
			var ec *string
			if errCode != "" {
				ec = &errCode
			}
			meta := map[string]any{}
			if errCode != "" {
				meta["response_error"] = errCode
			}
			e := domain.ApplicationErrorLog{
				ErrorType:   etype,
				ErrorCode:   ec,
				Message:     truncateStr(msg, 4000),
				HTTPStatus:  st,
				Method:      r.Method,
				Path:        r.URL.Path,
				QueryString: pqs,
				UserID:      uid,
				UserRole:    urole,
				RequestID:   rid,
				IPAddress:   pip,
				UserAgent:   pu,
				Meta:        meta,
			}
			go insertAppErrorAsync(inserter, e)
		})
	}
}

func insertAppErrorAsync(inserter ApplicationErrorLogInserter, e domain.ApplicationErrorLog) {
	defer func() { recover() }()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	_ = inserter.Insert(ctx, e)
}

func shouldSkipErrorLogPath(path string) bool {
	p := strings.TrimSuffix(path, "/")
	switch p {
	case "", "/api/v1/health", "/api/v1":
		return true
	}
	if strings.HasPrefix(p, "/api/v1/geo") {
		return true
	}
	return false
}

func parseErrorJSON(body []byte) (code string, message string) {
	var v struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		return "", strings.TrimSpace(string(body))
	}
	return strings.TrimSpace(v.Error), strings.TrimSpace(v.Message)
}

func inferApplicationErrorType(httpStatus int, errorCode string) string {
	ec := strings.ToLower(strings.TrimSpace(errorCode))
	switch {
	case strings.Contains(ec, "validation") || ec == "bad_request" || ec == "invalid_body":
		return domain.AppErrTypeValidation
	case ec == "unauthorized" || ec == "invalid_creds":
		return domain.AppErrTypeAuthentication
	case ec == "forbidden" || ec == "password_setup_required":
		return domain.AppErrTypeAuthorization
	case ec == "not_found":
		return domain.AppErrTypeNotFound
	case ec == "conflict":
		return domain.AppErrTypeConflict
	case strings.Contains(ec, "rate"):
		return domain.AppErrTypeRateLimit
	case ec == "service_unavailable":
		return domain.AppErrTypeExternal
	case ec == "internal_error" || ec == "server_error":
		return domain.AppErrTypeInternal
	}
	switch httpStatus {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return domain.AppErrTypeValidation
	case http.StatusUnauthorized:
		return domain.AppErrTypeAuthentication
	case http.StatusForbidden:
		return domain.AppErrTypeAuthorization
	case http.StatusNotFound:
		return domain.AppErrTypeNotFound
	case http.StatusConflict:
		return domain.AppErrTypeConflict
	case http.StatusTooManyRequests:
		return domain.AppErrTypeRateLimit
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return domain.AppErrTypeExternal
	case http.StatusInternalServerError:
		return domain.AppErrTypeInternal
	default:
		if httpStatus >= 500 {
			return domain.AppErrTypeServer
		}
		if httpStatus >= 400 {
			return domain.AppErrTypeClient
		}
		return domain.AppErrTypeClient
	}
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
