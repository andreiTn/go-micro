package event

import (
	"github.com/andreiTn/go-micro/pkg/amqpkit"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Emitter struct {
	connection *amqp.Connection
}

func NewEventEmitter(conn *amqp.Connection) (Emitter, error) {
	emitter := Emitter{connection: conn}
	if err := emitter.setup(); err != nil {
		return Emitter{}, err
	}
	return emitter, nil
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return amqpkit.DeclareTopicExchange(channel)
}

func (e *Emitter) Push(event, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	return channel.Publish(
		amqpkit.ExchangeLogsTopic,
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(event),
		},
	)
}
