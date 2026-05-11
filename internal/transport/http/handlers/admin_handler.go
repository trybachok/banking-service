package handlers

import (
	"net/http"
	"strconv"

	adminapp "github.com/example/banking-service/internal/application/admin"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AdminHandler struct {
	admin *adminapp.Service
	log   *logrus.Logger
}

func NewAdminHandler(admin *adminapp.Service, log *logrus.Logger) *AdminHandler {
	return &AdminHandler{
		admin: admin,
		log:   log,
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := parseQueryInt(r, "limit", 50)
	offset := parseQueryInt(r, "offset", 0)

	result, err := h.admin.ListUsers(r.Context(), limit, offset)
	if err != nil {
		h.log.WithError(err).Warn("admin list users failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result,
	})
}

func (h *AdminHandler) BlockAccount(w http.ResponseWriter, r *http.Request) {
	accountID := domain.AccountID(mux.Vars(r)["accountId"])

	var req adminapp.BlockAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.admin.BlockAccount(r.Context(), accountID, req.Reason)
	if err != nil {
		h.log.WithError(err).WithField("account_id", accountID.String()).Warn("admin block account failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithField("account_id", accountID.String()).Info("account blocked by admin")

	response.WriteJSON(w, http.StatusOK, result)
}

func parseQueryInt(r *http.Request, name string, defaultValue int) int {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return value
}
