package middleware

import (
	"context"
	"net/http"
	"strings"

	appjwt "github.com/example/banking-service/internal/infrastructure/jwt"
	"github.com/example/banking-service/internal/transport/http/response"
)

type authContextKey struct{}

type AuthUser struct {
	UserID string
	Role   string
}

func Auth(manager *appjwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))

			if header == "" {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authorization header is required")
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(header, prefix) {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authorization header must use Bearer scheme")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
			if token == "" {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "token is required")
				return
			}

			claims, err := manager.Parse(token)
			if err != nil {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "invalid token")
				return
			}

			authUser := AuthUser{
				UserID: claims.UserID,
				Role:   claims.Role,
			}

			ctx := context.WithValue(r.Context(), authContextKey{}, authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentUser(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(authContextKey{}).(AuthUser)
	return user, ok
}
