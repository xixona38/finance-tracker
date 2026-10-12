package server

import (
	"net/http"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/transaction"
	"github.com/xixona38/finance-tracker/internal/user"
)

// NewRouter connects the HTTP routes to their handlers and adds CSRF protection.
// Account routes, transaction routes, and /me require a valid session.
func NewRouter(
	acc *account.Handler,
	authHandler *auth.Handler,
	trHandler *transaction.Handler,
	userHandler *user.Handler,
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
	mux.Handle("POST /api/v1/transactions", mw.RequireAuth(http.HandlerFunc(trHandler.Create)))
	mux.Handle("GET /api/v1/transactions", mw.RequireAuth(http.HandlerFunc(trHandler.List)))
	mux.Handle("DELETE /api/v1/transactions/{id}", mw.RequireAuth(http.HandlerFunc(trHandler.Delete)))
	mux.Handle("GET /api/v1/me", mw.RequireAuth(http.HandlerFunc(userHandler.Me)))
	return secureMux
}
