package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/andreiTn/go-micro/logger-service/internal/data"
	"github.com/andreiTn/go-micro/pkg/config"
	"github.com/andreiTn/go-micro/pkg/httpserver"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

type Config struct {
	Models     data.Models
	Mongo      *mongo.Client
	RPCPort    string
	GRPCPort   string
	grpcServer *grpc.Server
	rpcListener net.Listener
}

func main() {
	level := slog.LevelInfo
	if config.GetOr("LOG_LEVEL", "info") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	mongoURL := config.GetOr("MONGO_URL", "mongodb://mongo:27017")
	mongoUser := config.GetOr("MONGO_USER", "admin")
	mongoPass := config.GetOr("MONGO_PASSWORD", "password")
	mongoDB := config.GetOr("MONGO_DB", "logs")
	mongoColl := config.GetOr("MONGO_COLLECTION", "logs")

	client, err := connectMongo(mongoURL, mongoUser, mongoPass, logger)
	if err != nil {
		logger.Error("mongo connection failed", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("mongo disconnect", "err", err)
		}
	}()

	port := config.GetOr("PORT", "80")
	rpcPort := config.GetOr("RPC_PORT", "5001")
	grpcPort := config.GetOr("GRPC_PORT", "50001")
	shutdownTimeout := config.Duration("SHUTDOWN_TIMEOUT", 10*time.Second)

	app := Config{
		Models:   data.New(client, mongoDB, mongoColl),
		Mongo:    client,
		RPCPort:  rpcPort,
		GRPCPort: grpcPort,
	}

	if err := app.startRPC(logger); err != nil {
		logger.Error("rpc listen failed", "err", err)
		os.Exit(1)
	}
	defer app.stopRPC(logger)

	if err := app.startGRPC(logger); err != nil {
		logger.Error("grpc listen failed", "err", err)
		os.Exit(1)
	}
	defer app.stopGRPC(logger)

	logger.Info("starting logger http service", "port", port)

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

func connectMongo(url, user, pass string, logger *slog.Logger) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(url)
	if user != "" {
		clientOptions.SetAuth(options.Credential{
			Username: user,
			Password: pass,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	logger.Info("connected to mongo")
	return client, nil
}

func (app *Config) startRPC(logger *slog.Logger) error {
	rpc.Register(&RPCServer{Models: app.Models})

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", app.RPCPort))
	if err != nil {
		return err
	}
	app.rpcListener = listener

	go func() {
		logger.Info("rpc server listening", "addr", listener.Addr().String())
		for {
			conn, err := listener.Accept()
			if err != nil {
				if app.rpcListener == nil {
					return
				}
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	return nil
}

func (app *Config) stopRPC(logger *slog.Logger) {
	if app.rpcListener != nil {
		_ = app.rpcListener.Close()
		app.rpcListener = nil
		logger.Info("rpc server stopped")
	}
}

func (app *Config) startGRPC(logger *slog.Logger) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", app.GRPCPort))
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	app.grpcServer = s

	registerLogService(s, &LogServer{Models: app.Models})

	go func() {
		logger.Info("grpc server listening", "addr", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			logger.Error("grpc serve", "err", err)
		}
	}()

	return nil
}

func (app *Config) stopGRPC(logger *slog.Logger) {
	if app.grpcServer != nil {
		app.grpcServer.GracefulStop()
		app.grpcServer = nil
		logger.Info("grpc server stopped")
	}
}

func (app *Config) routes(logger *slog.Logger) http.Handler {
	return app.buildRoutes(logger)
}
