package handlers

import (
	"net/http"

	cardapp "github.com/example/banking-service/internal/application/cards"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type CardHandler struct {
	cards *cardapp.Service
	log   *logrus.Logger
}

func NewCardHandler(cards *cardapp.Service, log *logrus.Logger) *CardHandler {
	return &CardHandler{
		cards: cards,
		log:   log,
	}
}

func (h *CardHandler) Issue(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	var req cardapp.IssueCardRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.cards.IssueCard(r.Context(), domain.UserID(currentUser.UserID), req)
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("issue card failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id": currentUser.UserID,
		"card_id": result.Card.ID,
	}).Info("card issued")

	response.WriteJSON(w, http.StatusCreated, result)
}

func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	result, err := h.cards.ListCards(r.Context(), domain.UserID(currentUser.UserID))
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("list cards failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result,
	})
}

func (h *CardHandler) Details(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	cardID := domain.CardID(mux.Vars(r)["cardId"])

	result, err := h.cards.GetCardDetails(r.Context(), domain.UserID(currentUser.UserID), cardID)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id": currentUser.UserID,
			"card_id": cardID.String(),
		}).Warn("get card details failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

func (h *CardHandler) Pay(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	cardID := domain.CardID(mux.Vars(r)["cardId"])

	var req cardapp.CardPaymentRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.cards.Pay(r.Context(), domain.UserID(currentUser.UserID), cardID, req)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id": currentUser.UserID,
			"card_id": cardID.String(),
		}).Warn("card payment failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":        currentUser.UserID,
		"card_id":        cardID.String(),
		"transaction_id": result.Transaction.ID,
	}).Info("card payment completed")

	response.WriteJSON(w, http.StatusOK, result)
}
