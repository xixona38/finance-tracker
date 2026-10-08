package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/platform/database"
	"github.com/xixona38/finance-tracker/internal/server"
	"github.com/xixona38/finance-tracker/internal/session"
	"github.com/xixona38/finance-tracker/internal/transaction"
	"github.com/xixona38/finance-tracker/internal/user"
)

// main connects to the database, sets up the application, and starts the HTTP server.
func main() {
	pgURL, ok := os.LookupEnv("PG_URL")
	if !ok || pgURL == "" {
		fmt.Println("There is no variable for PG_URL")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pool, err := database.NewPool(ctx, pgURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pool.Close()

	accRepo := account.NewRepository(pool)
	accSvc := account.NewService(accRepo)
	accHand := account.NewHandler(accSvc)

	sessionRepo := session.NewSessionRepo(pool)
	sessionSvc, err := session.NewService(sessionRepo, time.Minute*15, time.Hour*10)
	if err != nil {
		fmt.Println(err)
		return
	}

	userRepo := user.NewRepository(pool)

	authSvc := auth.NewService(userRepo, sessionSvc)
	authHand := auth.NewHandler(authSvc)
	mw := auth.NewMiddleware(sessionSvc)

	trRepo := transaction.NewRepository(pool)
	trSvc := transaction.NewService(trRepo, accRepo)
	trHand := transaction.NewHandler(trSvc)

	mux := server.NewRouter(accHand, authHand, trHand, mw)

	server := http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 5,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
