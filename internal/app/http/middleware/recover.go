package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/jsonerror"
	"github.com/meirusfandi/fansedu-golang-api/internal/domain"
)

func Recover(inserter ApplicationErrorLogInserter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic: %v", rec)
					stack := debug.Stack()
					s := string(stack)
					if inserter != nil && !shouldSkipErrorLogPath(r.URL.Path) {
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
						msg := fmt.Sprintf("%v", rec)
						e := domain.ApplicationErrorLog{
							ErrorType:   domain.AppErrTypePanic,
							Message:     truncateStr(msg, 4000),
							HTTPStatus:  http.StatusInternalServerError,
							Method:      r.Method,
							Path:        r.URL.Path,
							UserID:      uid,
							UserRole:    urole,
							RequestID:   rid,
							IPAddress:   pip,
							UserAgent:   pu,
							StackTrace:  &s,
							Meta:        map[string]any{"panic": true},
						}
						go func() {
							defer func() { recover() }()
							ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
							defer cancel()
							_ = inserter.Insert(ctx, e)
						}()
					}
					if strings.HasPrefix(r.URL.Path, "/api/") {
						jsonerror.Write(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada server.")
					} else {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
