package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/user"
)

type Handler struct {
	svc          *Service
	cookieSecure bool
	logger       *slog.Logger
}

// NewHandler sets up the registration, login, and logout handlers using the given service.
func NewHandler(svc *Service, cookieSecure bool, logger *slog.Logger) *Handler {
	return &Handler{
		svc:          svc,
		cookieSecure: cookieSecure,
		logger:       logger,
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

// Register reads an email and password from JSON and creates a user.
// It returns the user's public data with status 201, without the password hash.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	req := registerRequest{}
	var emptyVar any
	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&req); err != nil {
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

	registeredUser, err := h.svc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}

		if errors.Is(err, user.ErrMailAlreadyExists) {
			httpresponse.WriteError(w, http.StatusConflict, err.Error(), h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to register user",
			"error", err,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to register a user", h.logger)
		return
	}

	httpresponse.WriteJSON(w, http.StatusCreated, registeredUser, h.logger)

}

// Login checks the submitted email and password and starts a new session.
// It puts the session token in a cookie and returns the user's public data as JSON.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	req := loginRequest{}
	var emptyData any

	reader := http.MaxBytesReader(w, r.Body, 16*1024)
	decoder := json.NewDecoder(reader)

	if err := decoder.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error(), h.logger)
			return
		}

		httpresponse.WriteError(w, http.StatusBadRequest, "failed to decode json", h.logger)
		return
	}

	if err := decoder.Decode(&emptyData); err != io.EOF {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpresponse.WriteError(w, http.StatusRequestEntityTooLarge, maxBytesErr.Error(), h.logger)
			return
		}
		httpresponse.WriteError(w, http.StatusBadRequest, "too much data", h.logger)
		return
	}

	loginResult, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			httpresponse.WriteError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}

		if errors.Is(err, ErrInvalidCredentials) {
			httpresponse.WriteError(w, http.StatusUnauthorized, "wrong password or email", h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to sign in",
			"error", err,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to sign in", h.logger)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    loginResult.Token,
		Path:     "/",
		Expires:  loginResult.ExpiresAt,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	httpresponse.WriteJSON(w, http.StatusOK, loginResult.User, h.logger)
}

// Logout ends the session and tells the client to remove its session cookie.
// It returns status 204 with no body, including when the client has no session cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session")
	if err != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   h.cookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.svc.Logout(r.Context(), token.Value); err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"Failed to logout",
			"error", err,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "failed to logout", h.logger)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
