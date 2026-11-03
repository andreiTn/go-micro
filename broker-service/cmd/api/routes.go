package main

import (
	"log/slog"
	"net/http"

	"errors"

	"github.com/andreiTn/go-micro/pkg/httpx"
	"github.com/go-chi/chi/v5"
)

func (app *Config) routes(logger *slog.Logger) http.Handler {
	mux := chi.NewRouter()
	httpx.MountCommon(mux, logger)

	mux.Get("/healthz", httpx.Healthz)
	mux.Get("/readyz", httpx.Readyz(func(*http.Request) error {
		if app.Rabbit == nil || app.Rabbit.IsClosed() {
			return errors.New("rabbitmq connection closed")
		}
		ch, err := app.Rabbit.Channel()
		if err != nil {
			return err
		}
		return ch.Close()
	}))

	mux.Get("/", app.Broker)
	mux.Post("/", app.Broker)
	mux.Post("/handle", app.HandleSubmission)
	mux.Post("/log-grpc", app.LogViaGRPC)

	return mux
}
