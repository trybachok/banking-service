package middleware

import "net/http"

func RequireMFA() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}
