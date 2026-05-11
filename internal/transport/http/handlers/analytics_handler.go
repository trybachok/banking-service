package handlers

import (
	"net/http"
	"strconv"
	"time"

	analyticsapp "github.com/example/banking-service/internal/application/analytics"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type AnalyticsHandler struct {
	analytics *analyticsapp.Service
	log       *logrus.Logger
}

func NewAnalyticsHandler(analytics *analyticsapp.Service, log *logrus.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analytics: analytics,
		log:       log,
	}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	result, err := h.analytics.GetAnalytics(r.Context(), domain.UserID(currentUser.UserID), time.Now().UTC())
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("get analytics failed")
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}

func (h *AnalyticsHandler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	accountID := domain.AccountID(mux.Vars(r)["accountId"])

	daysRaw := r.URL.Query().Get("days")
	if daysRaw == "" {
		daysRaw = "30"
	}

	days, err := strconv.Atoi(daysRaw)
	if err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "validation_error", "days must be integer")
		return
	}

	result, err := h.analytics.PredictBalance(
		r.Context(),
		domain.UserID(currentUser.UserID),
		accountID,
		days,
		time.Now().UTC(),
	)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id":    currentUser.UserID,
			"account_id": accountID.String(),
			"days":       days,
		}).Warn("predict balance failed")

		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}
