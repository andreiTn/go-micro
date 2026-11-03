package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/andreiTn/go-micro/pkg/config"
	"github.com/andreiTn/go-micro/pkg/httpserver"
)

type pageConfig struct {
	BrokerURL string `json:"brokerUrl"`
}

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	port := config.GetOr("PORT", "80")
	brokerURL := config.GetOr("BROKER_URL", "http://localhost:8080")
	templatesDir := config.GetOr("TEMPLATES_DIR", "./cmd/web/templates")
	staticDir := config.GetOr("STATIC_DIR", "./cmd/web/static")
	shutdownTimeout := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)

	cfgJSON, err := json.Marshal(pageConfig{BrokerURL: brokerURL})
	if err != nil {
		logger.Error("marshal page config", "err", err)
		os.Exit(1)
	}

	tmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "base.layout.gohtml"),
		filepath.Join(templatesDir, "console.page.gohtml"),
	)
	if err != nil {
		logger.Error("parse templates", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "base", map[string]any{
			"ConfigJSON": template.JS(cfgJSON),
			"BrokerURL":  brokerURL,
		}); err != nil {
			logger.Error("render template", "err", err)
			http.Error(w, "template error", http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":false,"message":"ok"}`))
	})

	logger.Info("starting front-end", "port", port, "broker_url", brokerURL)
	if err := httpserver.Run(httpserver.Options{
		Addr:            fmt.Sprintf(":%s", port),
		Handler:         mux,
		ShutdownTimeout: shutdownTimeout,
		Logger:          logger,
	}); err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}
