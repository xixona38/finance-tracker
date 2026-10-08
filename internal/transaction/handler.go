package transaction

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

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
