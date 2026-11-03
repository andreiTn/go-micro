package main

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/andreiTn/go-micro/pkg/httpx"
	"github.com/go-chi/chi/v5"
)

func (app *Config) routes(logger *slog.Logger) http.Handler {
	mux := chi.NewRouter()
	httpx.MountCommon(mux, logger)

	mux.Get("/healthz", httpx.Healthz)
	mux.Get("/readyz", httpx.Readyz(func(*http.Request) error {
		return app.DB.PingContext(context.Background())
	}))

	mux.Post("/authenticate", app.authenticate)

	return mux
}
