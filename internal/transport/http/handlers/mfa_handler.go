package handlers

import (
	"net/http"

	mfaapp "github.com/example/banking-service/internal/application/mfa"
	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/transport/http/middleware"
	"github.com/example/banking-service/internal/transport/http/response"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type MFAHandler struct {
	mfa *mfaapp.Service
	log *logrus.Logger
}

func NewMFAHandler(mfa *mfaapp.Service, log *logrus.Logger) *MFAHandler {
	return &MFAHandler{
		mfa: mfa,
		log: log,
	}
}

func (h *MFAHandler) CreateChallenge(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	var req mfaapp.CreateChallengeRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.mfa.CreateChallenge(r.Context(), domain.UserID(currentUser.UserID), req)
	if err != nil {
		h.log.WithError(err).WithField("user_id", currentUser.UserID).Warn("create mfa challenge failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":      currentUser.UserID,
		"challenge_id": result.ID,
		"purpose":      result.Purpose,
	}).Info("mfa challenge created")

	response.WriteJSON(w, http.StatusCreated, result)
}

func (h *MFAHandler) VerifyChallenge(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := middleware.CurrentUser(r.Context())
	if !ok {
		response.WriteErrorMessage(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}

	challengeID := domain.MFAChallengeID(mux.Vars(r)["challengeId"])

	var req mfaapp.VerifyChallengeRequest
	if err := decodeJSON(r, &req); err != nil {
		response.WriteErrorMessage(w, http.StatusBadRequest, "invalid_json", "invalid request body")
		return
	}

	result, err := h.mfa.VerifyChallenge(r.Context(), domain.UserID(currentUser.UserID), challengeID, req)
	if err != nil {
		h.log.WithError(err).WithFields(logrus.Fields{
			"user_id":      currentUser.UserID,
			"challenge_id": challengeID.String(),
		}).Warn("verify mfa challenge failed")
		response.WriteError(w, err)
		return
	}

	h.log.WithFields(logrus.Fields{
		"user_id":      currentUser.UserID,
		"challenge_id": challengeID.String(),
	}).Info("mfa challenge verified")

	response.WriteJSON(w, http.StatusOK, result)
}
