package main

import (
	"context"
	"time"

	"github.com/andreiTn/go-micro/logger-service/internal/data"
)

type RPCServer struct {
	Models data.Models
}

type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := r.Models.LogEntry.Insert(ctx, data.LogEntry{
		Name: payload.Name,
		Data: payload.Data,
	})
	if err != nil {
		return err
	}

	*resp = "Processed payload via RPC " + payload.Name
	return nil
}
