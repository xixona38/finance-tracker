package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/xixona38/finance-tracker/internal/platform/httpresponse"
	"github.com/xixona38/finance-tracker/internal/session"
)

type userIDContextKey struct{}

type Middleware struct {
	ses *session.Service
}

// NewMiddleware creates authentication middleware with the supplied session service.
func NewMiddleware(session *session.Service) *Middleware {
	return &Middleware{
		ses: session,
	}
}

// RequireAuth validates the session cookie and passes the session's user ID to next through the request context.
// It returns 401 for a missing cookie, missing session, or expired session, and 500 for other validation errors.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			httpresponse.WriteError(w, http.StatusUnauthorized, "session token not found")
			return
		}

		sessionFound, err := m.ses.Validate(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, session.ErrSessionExpired) || errors.Is(err, session.ErrSessionNotFound) {
				httpresponse.WriteError(w, http.StatusUnauthorized, "session expired or not found")
				return
			}

			httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey{}, sessionFound.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
