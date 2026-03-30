package middleware

import (
	"net/http"
	"strings"
)

// corsAllowedHeaders — daftar eksplisit (bukan "*") karena browser modern (Chrome m97+)
// tidak lagi menganggap wildcard mencakup Authorization pada preflight.
const corsAllowedHeaders = "Accept, Accept-Language, Authorization, Content-Type, Origin, X-Requested-With, X-CSRF-Token, Cache-Control"

// CORS returns a middleware that sets CORS headers and handles OPTIONS preflight.
// Allowed origins: set CORS_ORIGINS env (comma-separated, e.g. "http://localhost:5173,https://app.example.com").
// Default allows http://localhost:5173 (Vite frontend) and "*".
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.TrimSpace(allowedOrigins)
	if origins == "" {
		origins = "http://localhost:5173,*"
	}
	allowAll := origins == "*"
	originList := strings.Split(origins, ",")
	for i := range originList {
		originList[i] = strings.TrimSpace(originList[i])
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				for _, o := range originList {
					if o == origin || o == "*" {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", corsAllowedHeaders)
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
