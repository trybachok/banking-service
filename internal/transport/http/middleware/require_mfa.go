package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/response"
)

type MFAVerifier interface {
	EnsureVerifiedChallenge(ctx context.Context, userID domain.UserID, purpose domain.MFAPurpose, challengeID domain.MFAChallengeID) error
}

func RequireMFA(verifier MFAVerifier, purpose domain.MFAPurpose) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentUser, ok := CurrentUser(r.Context())
			if !ok {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
				return
			}

			challengeID := strings.TrimSpace(r.Header.Get("X-MFA-Challenge-ID"))
			if challengeID == "" {
				response.WriteErrorMessage(w, http.StatusUnauthorized, "mfa_required", "X-MFA-Challenge-ID header is required")
				return
			}

			if err := verifier.EnsureVerifiedChallenge(
				r.Context(),
				domain.UserID(currentUser.UserID),
				purpose,
				domain.MFAChallengeID(challengeID),
			); err != nil {
				response.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
