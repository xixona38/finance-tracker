package transaction

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
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
	userID, ok := auth.UserIDFromContext(r.Context())
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
// It reads limit, offset, and the optional account_id filter from the URL.
// By default, it returns up to 20 transactions starting from the first result.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	var accID *int64

	lim := r.URL.Query().Get("limit")
	off := r.URL.Query().Get("offset")
	acc := r.URL.Query().Get("account_id")

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

	trs, err := h.svc.List(r.Context(), userID, accID, limit, offset)
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
