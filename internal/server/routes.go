package server

import (
	"net/http"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/transaction"
)

// NewRouter sets up the authentication, account, and transaction routes with CSRF protection.
// Account and transaction routes require a valid session.
func NewRouter(
	acc *account.Handler,
	authHandler *auth.Handler,
	trHandler *transaction.Handler,
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

	return secureMux
}
