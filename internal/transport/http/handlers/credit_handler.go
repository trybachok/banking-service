package handlers

import (
	"net/http"

	creditapp "github.com/example/banking-service/internal/application/credits"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type CreditHandler struct {
	credits *creditapp.Service
	log     *logrus.Logger
}

func NewCreditHandler(credits *creditapp.Service, log *logrus.Logger) *CreditHandler {
	return &CreditHandler{
		credits: credits,
		log:     log,
	}
}

func (h *CreditHandler) Issue(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	var req creditapp.IssueCreditRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.credits.IssueCredit(r.Context(), domain.UserID(currentUser.UserID), req)
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("issue credit failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":   currentUser.UserID,
		"credit_id": result.Credit.ID,
	}).Info("credit issued")

	response.WriteJSON(w, http.StatusCreated, result)
}

func (h *CreditHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	result, err := h.credits.ListCredits(r.Context(), domain.UserID(currentUser.UserID))
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("list credits failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result,
	})
}

func (h *CreditHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	creditID := domain.CreditID(mux.Vars(r)["creditId"])

	result, err := h.credits.GetSchedule(r.Context(), domain.UserID(currentUser.UserID), creditID)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id":   currentUser.UserID,
			"credit_id": creditID.String(),
		}).Warn("get credit schedule failed")

		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result,
	})
}
