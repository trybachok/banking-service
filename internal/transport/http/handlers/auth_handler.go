package handlers

import (
	"encoding/json"
	"net/http"

	authapp "github.com/example/banking-service/internal/application/auth"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	auth *authapp.Service
	log  *logrus.Logger
}

func NewAuthHandler(auth *authapp.Service, log *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		auth: auth,
		log:  log,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authapp.RegisterRequest

	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.auth.Register(r.Context(), req)
	if err != nil {
		h.log.WithError(err).Warn("register failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithField("user_id", result.User.ID).Info("user registered")

	response.WriteJSON(w, http.StatusCreated, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authapp.LoginRequest

	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), req)
	if err != nil {
		h.log.WithError(err).Warn("login failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithField("user_id", result.User.ID).Info("user authenticated")

	response.WriteJSON(w, http.StatusOK, result)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(target)
}
