package middleware

import (
	"net/http"

	"github.com/example/banking-service/internal/transport/http/response"
)

func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentUser, ok := CurrentUser(r.Context())
			if !ok {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}

			if currentUser.Role != requiredRole {
				response.WriteErrorMessage(w, http.StatusForbidden, "forbidden", "not enough permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
