package account

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

type accountDTO struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Currency       string `json:"currency"`
	InitialBalance int64  `json:"initial_balance"`
}

// NewHandler sets up the account handlers using the given account service.
func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// Create reads account details from JSON and creates an account for the logged-in user.
// It returns the new account with status 201, or an error if the request cannot be completed.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}

	accDTO := accountDTO{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&accDTO); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error(), h.logger)
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "failed to decode json", h.logger)
		return
	}

	if err := decoder.Decode(&emptyVar); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error(), h.logger)
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data", h.logger)
		return
	}
	acc := Account{
		UserID:         userID,
		Name:           accDTO.Name,
		Type:           accDTO.Type,
		Currency:       accDTO.Currency,
		InitialBalance: accDTO.InitialBalance,
	}
	createdAcc, err := h.svc.Create(r.Context(), acc)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		h.logger.ErrorContext(
			r.Context(),
			"Failed to create account",
			"error", err,
			"user_id", userID,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to create an account", h.logger)
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, createdAcc, h.logger)
}

// List returns the logged-in user's accounts and current balances as JSON.
// It reads the user ID from the request context, not from the URL.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}
	accounts, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"Failed to list accounts",
			"error", err,
			"user_id", userID,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to get accounts", h.logger)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, accounts, h.logger)
}
