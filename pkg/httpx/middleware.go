package httpx

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// DefaultCORS returns permissive CORS suitable for local demo front-ends.
func DefaultCORS() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// MountCommon applies CORS, request logging, recoverer, and /ping heartbeat.
func MountCommon(mux *chi.Mux, logger *slog.Logger) {
	mux.Use(DefaultCORS())
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(RequestLogger(logger))
	mux.Use(middleware.Heartbeat("/ping"))
}

// RequestLogger logs method, path, status, and duration.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
				"request_id", middleware.GetReqID(r.Context()),
			)
		})
	}
}

// Healthz is a liveness probe handler.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	_ = WriteJSON(w, http.StatusOK, Response{Message: "ok"})
}

// Readyz returns a readiness handler that calls check; nil check means always ready.
func Readyz(check func(*http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			if err := check(r); err != nil {
				_ = ErrorJSON(w, err, http.StatusServiceUnavailable)
				return
			}
		}
		_ = WriteJSON(w, http.StatusOK, Response{Message: "ready"})
	}
}
