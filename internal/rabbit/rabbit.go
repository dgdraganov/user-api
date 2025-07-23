package rabbit

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	channel *amqp.Channel
}

func NewRabbit(connStr string) (*Rabbit, error) {
	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("amqp dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("opening channel: %w", err)
	}

	return &Rabbit{
		channel: ch,
	}, nil
}

func (r *Rabbit) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	err := r.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	return nil
}
