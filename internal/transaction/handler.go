package transaction

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"
)

type Handler struct {
	svc *Service
}

type transactionDTO struct {
	FromAccID   *int64    `json:"from_acc_id"`
	ToAccID     *int64    `json:"to_acc_id"`
	TypeOp      string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// NewHandler sets up the transaction handlers using the given transaction service.
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Create reads transaction details from JSON and records the operation for the logged-in user.
// It returns the saved transaction with status 201, or an error if the request cannot be completed.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	trDTO := transactionDTO{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&trDTO); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error())
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "failed to decode json")
		return
	}

	if err := decoder.Decode(&emptyVar); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, "too much data")
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data")
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
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, account.ErrAccountNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "account not found")
			return
		}
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, res)
}

// List returns the logged-in user's transaction history as JSON.
// It reads limit, offset, and the optional account_id and type filters from the URL.
// By default, it returns up to 20 transactions starting from the first result.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required")
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
			httpresponse.WriteError(w, http.StatusBadRequest, "limit must be an integer")
			return
		}
		limit = val
	}

	if off != "" {
		val, err := strconv.Atoi(off)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, "offset must be an integer")
			return
		}
		offset = val
	}

	if acc != "" {
		val, err := strconv.ParseInt(acc, 10, 64)
		if err != nil {
			httpresponse.WriteError(w, http.StatusBadRequest, "account must be an integer")
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
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data")
			return
		}
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, trs)
}

// Delete reads the transaction ID from the URL path and deletes it for the logged-in user.
// It returns 204 with no body on success, or 404 if the transaction is missing or belongs to someone else.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	trIDStr := r.PathValue("id")
	if trIDStr == "" {
		httpresponse.WriteError(w, http.StatusBadRequest, "transaction id required")
		return
	}
	trId, err := strconv.ParseInt(trIDStr, 10, 64)
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "transaction id must be an integer")
		return
	}

	if err := h.svc.Delete(r.Context(), trId, userID); err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "transaction not found")
			return
		}
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data")
			return
		}
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
