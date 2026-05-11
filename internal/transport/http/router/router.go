package router

import (
	"net/http"

	"github.com/example/banking-service/internal/infrastructure/jwt"
	"github.com/example/banking-service/internal/transport/http/handlers"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func New(
	authHandler *handlers.AuthHandler,
	accountHandler *handlers.AccountHandler,
	transferHandler *handlers.TransferHandler,
	cardHandler *handlers.CardHandler,
	jwtManager *jwt.Manager,
	log *logrus.Logger,
) http.Handler {
	r := mux.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logging(log))

	r.HandleFunc("/health", healthHandler).Methods(http.MethodGet)
	r.HandleFunc("/ready", readyHandler).Methods(http.MethodGet)

	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	protected := r.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(jwtManager))

	protected.HandleFunc("/me", meHandler).Methods(http.MethodGet)

	protected.HandleFunc("/accounts", accountHandler.Create).Methods(http.MethodPost)
	protected.HandleFunc("/accounts", accountHandler.List).Methods(http.MethodGet)
	protected.HandleFunc("/accounts/{accountId}/deposit", accountHandler.Deposit).Methods(http.MethodPost)
	protected.HandleFunc("/accounts/{accountId}/withdraw", accountHandler.Withdraw).Methods(http.MethodPost)

	protected.HandleFunc("/transfer", transferHandler.Transfer).Methods(http.MethodPost)

	protected.HandleFunc("/cards", cardHandler.Issue).Methods(http.MethodPost)
	protected.HandleFunc("/cards", cardHandler.List).Methods(http.MethodGet)
	protected.HandleFunc("/cards/{cardId}", cardHandler.Details).Methods(http.MethodGet)
	protected.HandleFunc("/cards/{cardId}/pay", cardHandler.Pay).Methods(http.MethodPost)

	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "api",
	})
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"status":  "ready",
		"service": "api",
	})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{
		"userId": currentUser.UserID,
		"role":   currentUser.Role,
	})
}
