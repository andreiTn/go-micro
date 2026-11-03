package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/andreiTn/go-micro/pkg/httpx"
	"github.com/go-chi/chi/v5"
)

func (app *Config) buildRoutes(logger *slog.Logger) http.Handler {
	mux := chi.NewRouter()
	httpx.MountCommon(mux, logger)

	mux.Get("/healthz", httpx.Healthz)
	mux.Get("/readyz", httpx.Readyz(func(*http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return app.Mongo.Ping(ctx, nil)
	}))

	mux.Post("/log", app.WriteLog)

	return mux
}
