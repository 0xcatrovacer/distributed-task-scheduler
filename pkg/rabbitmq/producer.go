package rabbitmq

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	exchange   string
}

func NewProducer(amqpURL string, exchange string) (*Producer, error) {
	connection, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("error creating channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error declaring exchange: %w", err)
	}

	return &Producer{
		connection: connection,
		channel:    channel,
		exchange:   exchange,
	}, nil
}

func (p *Producer) PublishMessage(message string, routingKey string) error {
	err := p.channel.PublishWithContext(
		context.Background(),
		p.exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	return nil
}
