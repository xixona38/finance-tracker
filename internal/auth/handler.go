package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/user"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req := registerRequest{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&req); err != nil {
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

	registeredUser, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, user.ErrMailAlreadyExists) {
			httpresponse.WriteError(w, http.StatusConflict, err.Error())
			return
		}

		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to register a user")
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, registeredUser)

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req := loginRequest{}
	var emptyData any

	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error())
			return
		}

		httpresponse.WriteError(w, http.StatusBadRequest, "failed to decode json")
		return
	}

	if err := decoder.Decode(&emptyData); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error())
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data")
		return
	}

	authUser, err := h.svc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if errors.Is(err, ErrInvalidCredentials) {
			httpresponse.WriteError(w, http.StatusUnauthorized, "wrong password or email")
			return
		}

		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to sign in")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, authUser)
}
