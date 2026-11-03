package amqpkit

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeLogsTopic = "logs_topic"
	RoutingKeyInfo    = "log.INFO"
	RoutingKeyWarning = "log.WARNING"
	RoutingKeyError   = "log.ERROR"
)

// DefaultTopics are the routing keys the listener binds to.
var DefaultTopics = []string{RoutingKeyInfo, RoutingKeyWarning, RoutingKeyError}

// DialWithRetry connects to RabbitMQ with exponential backoff.
func DialWithRetry(url string, maxAttempts int, logger *slog.Logger) (*amqp.Connection, error) {
	if logger == nil {
		logger = slog.Default()
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			logger.Info("connected to rabbitmq")
			return conn, nil
		}
		lastErr = err
		backoff := time.Duration(math.Pow(float64(attempt), 2)) * time.Second
		logger.Warn("rabbitmq not ready", "attempt", attempt, "backoff", backoff.String(), "err", err)
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("rabbitmq dial failed after %d attempts: %w", maxAttempts, lastErr)
}

// DeclareTopicExchange declares the durable topic exchange used for logs.
func DeclareTopicExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		ExchangeLogsTopic,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

// DeclareEphemeralQueue declares an exclusive anonymous queue.
func DeclareEphemeralQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare("", false, false, true, false, nil)
}
