package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/platform/requestcontext"

	"github.com/xixona38/finance-tracker/internal/session"
)

type Middleware struct {
	ses          *session.Service
	cookieSecure bool
	logger       *slog.Logger
}

// NewMiddleware sets up session checks using the given session service.
// It stores the HTTPS-only cookie setting and a logger for unexpected session errors.
func NewMiddleware(session *session.Service, cookieSecure bool, logger *slog.Logger) *Middleware {
	return &Middleware{
		ses:          session,
		cookieSecure: cookieSecure,
		logger:       logger,
	}
}

// RequireAuth checks and renews the session before passing the request to the next handler.
// It refreshes the cookie's expiration time and adds the user's ID to the request context.
// A missing or expired session returns 401; other session errors are logged and return 500.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			httpresponse.WriteError(w, http.StatusUnauthorized, "session token not found", m.logger)
			return
		}

		sessionFound, err := m.ses.Renew(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, session.ErrSessionExpired) || errors.Is(err, session.ErrSessionNotFound) {
				httpresponse.WriteError(w, http.StatusUnauthorized, "session expired or not found", m.logger)
				return
			}

			m.logger.ErrorContext(
				r.Context(),
				"Failed to renew session",
				"error", err,
			)
			httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred", m.logger)
			return
		}

		ctx := requestcontext.WithUserID(r.Context(), sessionFound.UserID)

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    cookie.Value,
			Path:     "/",
			Expires:  sessionFound.ExpiresAt,
			HttpOnly: true,
			Secure:   m.cookieSecure,
			SameSite: http.SameSiteLaxMode,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
