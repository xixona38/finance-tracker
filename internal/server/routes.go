package server

import (
	"net/http"

	"github.com/xixona38/finance-tracker/internal/account"
)

func NewRouter(acc *account.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/accounts", acc.Create)
	return mux
}
