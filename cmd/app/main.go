package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xixona38/finance-tracker/internal/account"
	"github.com/xixona38/finance-tracker/internal/auth"
	"github.com/xixona38/finance-tracker/internal/config"
	"github.com/xixona38/finance-tracker/internal/platform/database"
	"github.com/xixona38/finance-tracker/internal/platform/httpmiddleware"
	"github.com/xixona38/finance-tracker/internal/server"
	"github.com/xixona38/finance-tracker/internal/session"
	"github.com/xixona38/finance-tracker/internal/transaction"
	"github.com/xixona38/finance-tracker/internal/user"
)

// main loads configuration, sets up logging and application dependencies, and starts the HTTP server.
// On shutdown, it gives active requests time to finish before closing the database pool.
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cnfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}

	dbCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	pool, err := database.NewPool(dbCtx, cnfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer pool.Close()

	rdbCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	rdb, err := database.NewRedisClient(rdbCtx, cnfg.RedisAddr)
	if err != nil {
		logger.Error("Failed to connect to redis", "error", err)
		return
	}
	defer rdb.Close()

	accRepo := account.NewRepository(pool)
	accSvc := account.NewService(accRepo)
	accHand := account.NewHandler(accSvc, logger)

	sessionRepo := session.NewRedisRepository(rdb)
	sessionSvc, err := session.NewService(sessionRepo, cnfg.SessionIdleTTL, cnfg.SessionAbsoluteTTL)
	if err != nil {
		logger.Error("Failed to create session service struct", "error", err)
		return
	}

	userRepo := user.NewRepository(pool)
	userSvc := user.NewService(userRepo)
	userHand := user.NewHandler(userSvc, logger)

	authSvc := auth.NewService(userRepo, sessionSvc)
	authHand := auth.NewHandler(authSvc, cnfg.CookieSecure, logger)
	mw := auth.NewMiddleware(sessionSvc, cnfg.CookieSecure, logger)

	trRepo := transaction.NewRepository(pool)
	trSvc := transaction.NewService(trRepo, accRepo)
	trHand := transaction.NewHandler(trSvc, logger)

	mux := server.NewRouter(accHand, authHand, trHand, userHand, mw)

	// Wrap the whole router so rejected requests and unknown routes are logged too.
	handler := httpmiddleware.Logging(logger, mux)

	server := http.Server{
		Addr:              cnfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: time.Second * 5,
	}

	// Turn Ctrl+C and SIGTERM into a cancellation that main can wait for.
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Keep room for one result so the server goroutine can finish even during shutdown.
	errChan := make(chan error, 1)

	logger.Info("Starting HTTP server", "addr", cnfg.HTTPAddr)
	// Serve requests while main waits for a signal or a server error.
	go func() {
		err := server.ListenAndServe()
		errChan <- err
	}()

	select {
	case <-sigCtx.Done():
		logger.Info("Shutdown signal received")

	case err := <-errChan:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", "error", err)
		}
	}

	// Restore normal signal handling so another Ctrl+C can interrupt shutdown.
	stop()

	// Start a fresh timeout: the signal context is already canceled after Ctrl+C.
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelShutdown()

	// Stop accepting new connections and wait up to ten seconds for active requests.
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)

		// A failed Shutdown does not force active connections to close.
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("Failed to close HTTP server", "error", closeErr)
		}

		return
	}

	logger.Info("HTTP server stopped")
}
