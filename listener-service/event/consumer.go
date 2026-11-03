package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/andreiTn/go-micro/pkg/amqpkit"
	"github.com/andreiTn/go-micro/pkg/clients"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn       *amqp.Connection
	loggerURL  string
	httpClient *clients.HTTP
}

func NewConsumer(conn *amqp.Connection, loggerURL string, httpClient *clients.HTTP) (Consumer, error) {
	consumer := Consumer{
		conn:       conn,
		loggerURL:  loggerURL,
		httpClient: httpClient,
	}

	if err := consumer.setup(); err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return amqpkit.DeclareTopicExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(ctx context.Context, topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := amqpkit.DeclareEphemeralQueue(ch)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if err := ch.QueueBind(q.Name, topic, amqpkit.ExchangeLogsTopic, false, nil); err != nil {
			return err
		}
	}

	deliveries, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	slog.Info("waiting for messages", "exchange", amqpkit.ExchangeLogsTopic, "queue", q.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return nil
			}
			var payload Payload
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				slog.Warn("invalid payload", "err", err)
				continue
			}
			go func(p Payload) {
				if err := consumer.handlePayload(ctx, p); err != nil {
					slog.Error("handle payload", "err", err)
				}
			}(payload)
		}
	}
}

func (consumer *Consumer) handlePayload(ctx context.Context, payload Payload) error {
	switch payload.Name {
	case "log", "event", "auth":
		return consumer.logEvent(ctx, payload)
	default:
		return consumer.logEvent(ctx, payload)
	}
}

func (consumer *Consumer) logEvent(ctx context.Context, entry Payload) error {
	_, err := consumer.httpClient.PostJSONExpect(ctx, consumer.loggerURL, entry, http.StatusAccepted)
	if err != nil {
		return fmt.Errorf("log event: %w", err)
	}
	return nil
}
