package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestIDContextKey struct{}

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			w.Header().Set("X-Request-ID", requestID)

			ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDContextKey{}).(string)
	return value
}

func generateRequestID() string {
	var buffer [16]byte

	if _, err := rand.Read(buffer[:]); err != nil {
		return "request-id-unavailable"
	}

	return hex.EncodeToString(buffer[:])
}
