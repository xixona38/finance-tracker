package server

import (
	"net/http"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
)

// NewRouter sets up the registration, login, logout, and account routes with CSRF protection.
// Account routes also require a valid session.
func NewRouter(
	acc *account.Handler,
	authHandler *auth.Handler,
	mw *auth.Middleware,
) http.Handler {
	protection := http.NewCrossOriginProtection()
	mux := http.NewServeMux()
	secureMux := protection.Handler(mux)

	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/logout", authHandler.Logout)
	mux.Handle("POST /api/v1/accounts", mw.RequireAuth(http.HandlerFunc(acc.Create)))
	mux.Handle("GET /api/v1/accounts", mw.RequireAuth(http.HandlerFunc(acc.List)))

	return secureMux
}
