package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/config"
	"github.com/xixona38/finance-tracker/internal/platform/database"
	"github.com/xixona38/finance-tracker/internal/server"
	"github.com/xixona38/finance-tracker/internal/session"
	"github.com/xixona38/finance-tracker/internal/transaction"
	"github.com/xixona38/finance-tracker/internal/user"
)

// main connects to the database, sets up the application, and starts the HTTP server.
// On shutdown, it gives active requests time to finish before closing the database pool.
func main() {
	cnfg, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pool, err := database.NewPool(ctx, cnfg.DatabaseURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer pool.Close()

	accRepo := account.NewRepository(pool)
	accSvc := account.NewService(accRepo)
	accHand := account.NewHandler(accSvc)

	sessionRepo := session.NewSessionRepo(pool)
	sessionSvc, err := session.NewService(sessionRepo, cnfg.SessionIdleTTL, cnfg.SessionAbsoluteTTL)
	if err != nil {
		fmt.Println(err)
		return
	}

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHand := user.NewHandler(userSvc)

	authSvc := auth.NewService(userRepo, sessionSvc)
	authHand := auth.NewHandler(authSvc, cnfg.CookieSecure)
	mw := auth.NewMiddleware(sessionSvc, cnfg.CookieSecure)

	trRepo := transaction.NewRepository(pool)
	trSvc := transaction.NewService(trRepo, accRepo)
	trHand := transaction.NewHandler(trSvc)

	mux := server.NewRouter(accHand, authHand, trHand, userHand, mw)

	server := http.Server{
		Addr:              cnfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 5,
	}

	// Turn Ctrl+C and SIGTERM into a cancellation that main can wait for.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Keep room for one result so the server goroutine can finish even during shutdown.
	errChan := make(chan error, 1)

	// Serve requests while main waits for a signal or a server error.
	go func() {
		err := server.ListenAndServe()
		errChan <- err
	}()

	select {
	case <-sigCtx.Done():
		fmt.Print("shutdown signal received\n")

	case err := <-errChan:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("HTTP server error:", err)
		}
	}

	// Restore normal signal handling so another Ctrl+C can interrupt shutdown.
	stop()

	// Start a fresh timeout: the signal context is already canceled after Ctrl+C.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelShutdown()

	// Stop accepting new connections and wait up to ten seconds for active requests.
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("Graceful shutdown failed:", err)

		// A failed Shutdown does not force active connections to close.
		if closeErr := server.Close(); closeErr != nil {
			fmt.Println("Failed to close HTTP server:", closeErr)
		}

		return
	}

	fmt.Println("HTTP server stopped")
}
