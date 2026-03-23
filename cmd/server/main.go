package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gin-rocket/internal/bootstrap"

	"go.uber.org/zap"
)

func main() {
	app, err := bootstrap.NewApplication()
	if err != nil {
		log.Fatalf("bootstrap application: %v", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		app.Logger.Info("shutdown signal received")
	case err := <-serverErrCh:
		app.Logger.Error("http server exited unexpectedly", zap.Error(err))
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.Server.ShutdownTimeout)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("graceful shutdown failed", zap.Error(err))
		os.Exit(1)
	}
}
