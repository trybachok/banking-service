package handlers

import (
	"net/http"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AccountHandler struct {
	accounts *accountapp.Service
	log      *logrus.Logger
}

func NewAccountHandler(accounts *accountapp.Service, log *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accounts: accounts,
		log:      log,
	}
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	var req accountapp.CreateAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.accounts.CreateAccount(r.Context(), domain.UserID(currentUser.UserID), req)
	if err != nil {
		h.log.WithError(err).Warn("create account failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":    currentUser.UserID,
		"account_id": result.ID,
	}).Info("account created")

	response.WriteJSON(w, http.StatusCreated, result)
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	result, err := h.accounts.ListAccounts(r.Context(), domain.UserID(currentUser.UserID))
	if err != nil {
		h.log.WithError(err).Warn("list accounts failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"items": result,
	})
}

func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	h.moneyOperation(w, r, "deposit")
}

func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	h.moneyOperation(w, r, "withdraw")
}

func (h *AccountHandler) moneyOperation(w http.ResponseWriter, r *http.Request, operation string) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	accountID := mux.Vars(r)["accountId"]

	var req accountapp.MoneyOperationRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	req.AccountID = accountID

	var (
		result *accountapp.AccountOperationResponse
		err    error
	)

	switch operation {
	case "deposit":
		result, err = h.accounts.Deposit(r.Context(), domain.UserID(currentUser.UserID), req)
	case "withdraw":
		result, err = h.accounts.Withdraw(r.Context(), domain.UserID(currentUser.UserID), req)
	default:
		response.WriteErrorMessage(w, http.StatusBadRequest, "unsupported_operation", "unsupported operation")
		return
	}

	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id":    currentUser.UserID,
			"account_id": accountID,
			"operation":  operation,
		}).Warn("account operation failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":        currentUser.UserID,
		"account_id":     accountID,
		"operation":      operation,
		"transaction_id": result.Transaction.ID,
	}).Info("account operation completed")

	response.WriteJSON(w, http.StatusOK, result)
}
