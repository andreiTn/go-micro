package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Options configure graceful HTTP serving.
type Options struct {
	Addr            string
	Handler         http.Handler
	ShutdownTimeout time.Duration
	Logger          *slog.Logger
}

// Run starts the HTTP server and shuts down on SIGINT/SIGTERM.
func Run(opts Options) error {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.ShutdownTimeout <= 0 {
		opts.ShutdownTimeout = 10 * time.Second
	}
	if opts.Addr == "" {
		opts.Addr = ":80"
	}

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           opts.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		opts.Logger.Info("http server listening", "addr", opts.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		opts.Logger.Info("shutdown signal received", "signal", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), opts.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}
	opts.Logger.Info("http server stopped")
	return nil
}
