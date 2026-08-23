package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type componentResult struct {
	name string
	err  error
}

func (app *Application) Run(ctx context.Context) error {
	server := app.newHTTPServer()
	results := make(chan componentResult, 2)
	go func() {
		app.Logger.Info("suggestion handling API started", zap.String("address", app.Config.HTTPAddress))
		results <- componentResult{name: "http", err: server.ListenAndServe()}
	}()
	go func() {
		results <- componentResult{name: "notification-outbox", err: app.Worker.Run(ctx, app.Config.WorkerInterval)}
	}()
	var first componentResult
	select {
	case <-ctx.Done():
		first = componentResult{name: "signal", err: ctx.Err()}
	case first = <-results:
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), app.Config.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	if first.err != nil && !errors.Is(first.err, context.Canceled) && !errors.Is(first.err, http.ErrServerClosed) {
		return fmt.Errorf("%s component failed: %w", first.name, first.err)
	}
	select {
	case result := <-results:
		if result.err != nil && !errors.Is(result.err, context.Canceled) && !errors.Is(result.err, http.ErrServerClosed) {
			return fmt.Errorf("%s component failed during shutdown: %w", result.name, result.err)
		}
	case <-shutdownCtx.Done():
		return fmt.Errorf("graceful shutdown exceeded %s", app.Config.ShutdownTimeout)
	}
	return nil
}

func (app *Application) newHTTPServer() *http.Server {
	return &http.Server{
		Addr:              app.Config.HTTPAddress,
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       app.Config.RequestTimeout + 2*time.Second,
		WriteTimeout:      app.Config.RequestTimeout + 10*time.Second,
		IdleTimeout:       75 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
