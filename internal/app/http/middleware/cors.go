package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a middleware that sets CORS headers and handles OPTIONS preflight.
// Allowed origins: set CORS_ORIGINS env (e.g. "*" or "https://app.example.com,http://localhost:3000"). Default "*".
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	origins := strings.TrimSpace(allowedOrigins)
	if origins == "" {
		origins = "*"
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
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
