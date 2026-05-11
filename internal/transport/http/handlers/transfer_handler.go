package handlers

import (
	"net/http"

	transferapp "github.com/example/banking-service/internal/application/transfers"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/sirupsen/logrus"
)

type TransferHandler struct {
	transfers *transferapp.Service
	log       *logrus.Logger
}

func NewTransferHandler(transfers *transferapp.Service, log *logrus.Logger) *TransferHandler {
	return &TransferHandler{
		transfers: transfers,
		log:       log,
	}
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	var req transferapp.TransferRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.transfers.Transfer(r.Context(), domain.UserID(currentUser.UserID), req)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id":                currentUser.UserID,
			"source_account_id":      req.SourceAccountID,
			"destination_account_id": req.DestinationAccountID,
		}).Warn("transfer failed")

		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":                currentUser.UserID,
		"source_account_id":      req.SourceAccountID,
		"destination_account_id": req.DestinationAccountID,
		"transaction_id":         result.Transaction.ID,
	}).Info("transfer completed")

	response.WriteJSON(w, http.StatusOK, result)
}
