package server

import (
	"net/http"

	"github.com/xixona38/finance-tracker/internal/account"
)

// NewRouter creates an HTTP router mapping POST /api/v1/accounts to account creation.
func NewRouter(acc *account.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/accounts", acc.Create)
	return mux
}
