package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/andreiTn/go-micro/pkg/config"
	"github.com/andreiTn/go-micro/pkg/httpserver"
)

type Config struct {
	Mailer Mail
}

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	templatesDir := config.GetOr("TEMPLATES_DIR", "./templates")
	port := config.GetOr("PORT", "80")
	shutdownTimeout := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)

	app := Config{
		Mailer: createMail(templatesDir),
	}

	logger.Info("starting mail service", "port", port, "templates", templatesDir)

	if err := httpserver.Run(httpserver.Options{
		Addr:            fmt.Sprintf(":%s", port),
		Handler:         app.routes(logger),
		ShutdownTimeout: shutdownTimeout,
		Logger:          logger,
	}); err != nil {
		logger.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

func createMail(templatesDir string) Mail {
	return Mail{
		Domain:        config.Get("MAIL_DOMAIN"),
		Host:          config.Get("MAIL_HOST"),
		Port:          config.Int("MAIL_SMTP_PORT", config.Int("MAIL_PORT", 1025)),
		Username:      config.Get("MAIL_USERNAME"),
		Password:      config.Get("MAIL_PASSWORD"),
		Encryption:    config.Get("MAIL_ENCRYPTION"),
		FromName:      config.Get("MAIL_FROM_NAME"),
		FromAddress:   config.Get("MAIL_FROM_ADDRESS"),
		TemplatesDir:  templatesDir,
		htmlTemplate:  filepath.Join(templatesDir, "mail.html.gohtml"),
		plainTemplate: filepath.Join(templatesDir, "mail.plain.gohtml"),
	}
}
