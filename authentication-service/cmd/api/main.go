package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/andreiTn/go-micro/authentication-service/data"
	"github.com/andreiTn/go-micro/pkg/clients"
	"github.com/andreiTn/go-micro/pkg/config"
	"github.com/andreiTn/go-micro/pkg/httpserver"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Config struct {
	DB         *sql.DB
	Users      UserStore
	LoggerHTTP *clients.HTTP
	LoggerURL  string
}

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	logger.Info("starting authentication service")

	dsn := config.MustGet("DSN")
	maxRetries := config.Int("DB_MAX_RETRIES", 10)
	conn, err := connectDB(dsn, maxRetries, logger)
	if err != nil {
		logger.Error("postgres connection failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	httpTimeout := config.Duration("HTTP_CLIENT_TIMEOUT", 10*time.Second)
	loggerURL := config.GetOr("LOGGER_HTTP_URL", "http://logger-service/log")
	port := config.GetOr("PORT", "80")
	shutdownTimeout := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)

	app := Config{
		DB:         conn,
		Users:      data.New(conn),
		LoggerHTTP: clients.NewHTTP(httpTimeout),
		LoggerURL:  loggerURL,
	}

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

func connectDB(dsn string, maxRetries int, logger *slog.Logger) (*sql.DB, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			lastErr = err
		} else if err = db.Ping(); err != nil {
			_ = db.Close()
			lastErr = err
		} else {
			logger.Info("connected to postgres")
			return db, nil
		}
		logger.Warn("postgres not ready", "attempt", attempt, "err", lastErr)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("postgres dial failed after %d attempts: %w", maxRetries, lastErr)
}
