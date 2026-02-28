package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			dur := time.Since(start)

			rid, _ := GetRequestID(r.Context())
			log.Printf("method=%s path=%s rid=%s dur=%s", r.Method, r.URL.Path, rid, dur)
		})
	}
}

