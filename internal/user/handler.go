package user

import (
	"errors"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"
)

type Handler struct {
	svc *Service
}

// NewHandler sets up user handlers using the given user service.
func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// Me reads the logged-in user's ID from the request context and returns their public details as JSON.
// It returns 401 if the context has no user ID, or 404 if that user no longer exists.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data")
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, user)
}
