package main

import (
	"context"

	"github.com/andreiTn/go-micro/contracts/logs"
	"github.com/andreiTn/go-micro/logger-service/internal/data"

	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func registerLogService(s *grpc.Server, srv *LogServer) {
	logs.RegisterLogServiceServer(s, srv)
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Name: input.GetName(),
		Data: input.GetData(),
	}

	if err := l.Models.LogEntry.Insert(ctx, logEntry); err != nil {
		return &logs.LogResponse{Result: "Failed"}, err
	}

	return &logs.LogResponse{Result: "logged!"}, nil
}
