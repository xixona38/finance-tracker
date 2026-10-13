package user

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

// NewHandler sets up user handlers using the given user service.
func NewHandler(svc *Service, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// Me reads the logged-in user's ID from the request context and returns their public details as JSON.
// It returns 401 if the context has no user ID, or 404 if that user no longer exists.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := requestcontext.UserIDFromContext(r.Context())
	if !ok {
		httpresponse.WriteError(w, http.StatusUnauthorized, "authentication required", h.logger)
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrInvalidData) {
			httpresponse.WriteError(w, http.StatusBadRequest, "invalid data", h.logger)
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			httpresponse.WriteError(w, http.StatusNotFound, "user not found", h.logger)
			return
		}

		h.logger.ErrorContext(
			r.Context(),
			"Failed to get user details",
			"error", err,
			"user_id", userID,
		)
		httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred", h.logger)
		return
	}

	httpresponse.WriteJSON(w, http.StatusOK, user, h.logger)
}
