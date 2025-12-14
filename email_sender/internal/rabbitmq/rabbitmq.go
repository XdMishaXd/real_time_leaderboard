package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(url string) (*RabbitMQ, error) {
	const op = "rabbitmq.New"

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

func (r *RabbitMQ) StartReading(ctx context.Context, queueName string, handler func([]byte)) error {
	const op = "rabbitmq.StartReading"

	msgs, err := r.channel.Consume(
		queueName,
		"",    // consumer name
		false, // auto-ack выключен (чтобы подтверждать вручную)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			handler(msg.Body)
			_ = msg.Ack(false)
		}
	}
}

func (r *RabbitMQ) Close() {
	_ = r.channel.Close()
	_ = r.conn.Close()
}
