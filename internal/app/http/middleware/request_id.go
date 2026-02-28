package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type ctxKeyRequestID struct{}

const headerRequestID = "X-Request-Id"

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(headerRequestID)
			if rid == "" {
				rid = newRequestID()
			}
			w.Header().Set(headerRequestID, rid)
			ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyRequestID{})
	s, ok := v.(string)
	return s, ok
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

