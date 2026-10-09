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

// NewMiddleware sets up session checks using the given session service.
func NewMiddleware(session *session.Service) *Middleware {
	return &Middleware{
		ses: session,
	}
}

// RequireAuth checks and renews the session before passing the request to the next handler.
// It refreshes the cookie's expiration time and adds the user's ID to the request context.
// A missing or expired session returns 401; other session errors return 500.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil {
			httpresponse.WriteError(w, http.StatusUnauthorized, "session token not found")
			return
		}

		sessionFound, err := m.ses.Renew(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, session.ErrSessionExpired) || errors.Is(err, session.ErrSessionNotFound) {
				httpresponse.WriteError(w, http.StatusUnauthorized, "session expired or not found")
				return
			}

			httpresponse.WriteError(w, http.StatusInternalServerError, "an error occurred")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey{}, sessionFound.UserID)

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    cookie.Value,
			Path:     "/",
			Expires:  sessionFound.ExpiresAt,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
