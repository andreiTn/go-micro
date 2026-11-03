package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreiTn/go-micro/listener-service/event"
	"github.com/andreiTn/go-micro/pkg/amqpkit"
	"github.com/andreiTn/go-micro/pkg/clients"
	"github.com/andreiTn/go-micro/pkg/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	rabbitURL := config.GetOr("RABBITMQ_URL", "amqp://guest:guest@rabbitmq")
	maxRetries := config.Int("RABBIT_MAX_RETRIES", 5)
	loggerURL := config.GetOr("LOGGER_HTTP_URL", "http://logger-service/log")
	httpTimeout := config.Duration("HTTP_CLIENT_TIMEOUT", 10*time.Second)

	conn, err := amqpkit.DialWithRetry(rabbitURL, maxRetries, logger)
	if err != nil {
		logger.Error("rabbitmq connection failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	consumer, err := event.NewConsumer(conn, loggerURL, clients.NewHTTP(httpTimeout))
	if err != nil {
		logger.Error("consumer setup failed", "err", err)
		os.Exit(1)
	}

	logger.Info("listening for rabbitmq messages")

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumer.Listen(ctx, amqpkit.DefaultTopics)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && err != amqp.ErrClosed {
			logger.Error("consumer stopped", "err", err)
			os.Exit(1)
		}
	}
}
