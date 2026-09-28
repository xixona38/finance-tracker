package account

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
)

type Handler struct {
	svc *Service
}

type accountDTO struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Currency       string `json:"currency"`
	InitialBalance int64  `json:"initial_balance"`
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	accDTO := accountDTO{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&accDTO); err != nil {
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
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error())
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data")
		return
	}
	acc := Account{
		Name:           accDTO.Name,
		Type:           accDTO.Type,
		Currency:       accDTO.Currency,
		InitialBalance: accDTO.InitialBalance,
	}
	createdAcc, err := h.svc.Create(r.Context(), acc)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to create an account")
		return
	}
	httpresponse.WriteJSON(w, http.StatusCreated, createdAcc)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil {
		httpresponse.WriteError(w, http.StatusBadRequest, "user id required")
		return
	}
	accounts, err := h.svc.List(r.Context(), int64(id))
	if err != nil {
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to get accounts")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, accounts)
}
