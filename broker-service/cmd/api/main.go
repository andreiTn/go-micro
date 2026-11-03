package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/andreiTn/go-micro/contracts/logs"
	"github.com/andreiTn/go-micro/pkg/amqpkit"
	"github.com/andreiTn/go-micro/pkg/clients"
	"github.com/andreiTn/go-micro/pkg/config"
	"github.com/andreiTn/go-micro/pkg/httpserver"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Rabbit         *amqp.Connection
	AuthURL        string
	LoggerHTTPURL  string
	LoggerRPCAddr  string
	LoggerGRPCAddr string
	MailURL        string
	HTTP           *clients.HTTP
	LogGRPC        logs.LogServiceClient
	grpcConn       *grpc.ClientConn
}

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	rabbitURL := config.GetOr("RABBITMQ_URL", "amqp://guest:guest@rabbitmq")
	maxRetries := config.Int("RABBIT_MAX_RETRIES", 5)

	conn, err := amqpkit.DialWithRetry(rabbitURL, maxRetries, logger)
	if err != nil {
		logger.Error("rabbitmq connection failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()

	httpTimeout := config.Duration("HTTP_CLIENT_TIMEOUT", 10*time.Second)
	port := config.GetOr("PORT", "80")
	shutdownTimeout := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)
	grpcAddr := config.GetOr("LOGGER_GRPC_ADDR", "logger-service:50001")

	grpcConn, err := dialLoggerGRPC(grpcAddr)
	if err != nil {
		logger.Error("grpc dial failed", "addr", grpcAddr, "err", err)
		os.Exit(1)
	}
	defer grpcConn.Close()

	app := Config{
		Rabbit:         conn,
		AuthURL:        config.GetOr("AUTH_URL", "http://authentication-service/authenticate"),
		LoggerHTTPURL:  config.GetOr("LOGGER_HTTP_URL", "http://logger-service/log"),
		LoggerRPCAddr:  config.GetOr("LOGGER_RPC_ADDR", "logger-service:5001"),
		LoggerGRPCAddr: grpcAddr,
		MailURL:        config.GetOr("MAIL_URL", "http://mail-service/send"),
		HTTP:           clients.NewHTTP(httpTimeout),
		LogGRPC:        logs.NewLogServiceClient(grpcConn),
		grpcConn:       grpcConn,
	}

	logger.Info("starting broker service", "port", port)

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

// dialLoggerGRPC opens a long-lived client connection.
// passthrough:/// skips the DNS resolver (avoids multi-second IPv6 stalls on Docker Desktop).
func dialLoggerGRPC(addr string) (*grpc.ClientConn, error) {
	target := addr
	if !strings.Contains(addr, "://") {
		target = "passthrough:///" + addr
	}
	return grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
