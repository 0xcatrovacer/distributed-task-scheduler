package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queue      string
}

func NewProducer(amqpURL string, queueName string) (*Producer, error) {
	connection, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("error connecting to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("error creating channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error declaring queue: %w", err)
	}

	return &Producer{
		connection: connection,
		channel:    channel,
		queue:      queueName,
	}, nil
}
