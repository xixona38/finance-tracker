package transaction

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

type transactionDTO struct {
	FromAccID   *int64    `json:"from_acc_id"`
	ToAccID     *int64    `json:"to_acc_id"`
	TypeOp      string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// NewHandler sets up transaction handlers with a service and a logger for internal errors.
func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// Create reads transaction details from JSON and records the operation for the logged-in user.
// It returns the saved transaction with status 201, or an error if the request cannot be completed.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}

	trDTO := transactionDTO{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&trDTO); err != nil {
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
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, "too much data", h.logger)
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data", h.logger)
		return
	}

	tr := Transaction{
		UserID:      userID,
		FromAccID:   trDTO.FromAccID,
		ToAccID:     trDTO.ToAccID,
		Type:        trDTO.TypeOp,
		Amount:      trDTO.Amount,
		Description: trDTO.Description,
		OccurredAt:  trDTO.OccurredAt,
	}

	res, err := h.svc.Create(r.Context(), &tr)
	if err != nil {
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		if errors.Is(err, account.ErrAccountNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "account not found", h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to create a new transaction",
			"error", err,
			"user_id", userID,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred", h.logger)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, res, h.logger)
}

// List returns the logged-in user's transaction history as JSON.
// It reads limit, offset, and the optional account_id and type filters from the URL.
// By default, it returns up to 20 transactions starting from the first result.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}
	var accID *int64

	lim := r.URL.Query().Get("limit")
	off := r.URL.Query().Get("offset")
	acc := r.URL.Query().Get("account_id")
	typeOfOp := r.URL.Query().Get("type")

	limit := 20
	offset := 0
	accID = nil

	if lim != "" {
		val, err := strconv.Atoi(lim)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, "limit must be an integer", h.logger)
			return
		}
		limit = val
	}

	if off != "" {
		val, err := strconv.Atoi(off)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, "offset must be an integer", h.logger)
			return
		}
		offset = val
	}

	if acc != "" {
		val, err := strconv.ParseInt(acc, 10, 64)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, "account must be an integer", h.logger)
			return
		}
		accID = &val
	}

	filter := ListFilter{
		AccountID: accID,
		Type:      typeOfOp,
		Limit:     limit,
		Offset:    offset,
	}

	trs, err := h.svc.List(r.Context(), userID, filter)
	if err != nil {
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data", h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to list user's transactions",
			"error", err,
			"user_id", userID,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred", h.logger)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, trs, h.logger)
}

// Delete reads the transaction ID from the URL path and deletes it for the logged-in user.
// It returns 204 with no body on success, or 404 if the transaction is missing or belongs to someone else.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}

	trIDStr := r.PathValue("id")
	if trIDStr == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "transaction id required", h.logger)
		return
	}
	trId, err := strconv.ParseInt(trIDStr, 10, 64)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "transaction id must be an integer", h.logger)
		return
	}

	if err := h.svc.Delete(r.Context(), trId, userID); err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "transaction not found", h.logger)
			return
		}
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data", h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to delete user's transaction",
			"error", err,
			"user_id", userID,
			"transaction_id", trId,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred", h.logger)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
