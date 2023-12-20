package rabbitmq

import (
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queue      string
}

func NewConsumer(amqpURL string, queueName string) (*Consumer, error) {
	connection, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
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
		return nil, fmt.Errorf("unable to declare a queue: %w", err)
	}

	return &Consumer{
		connection: connection,
		channel:    channel,
		queue:      queueName,
	}, nil
}

func (c *Consumer) Consume() {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err.Error())
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			msg, err := ParseMessage(d.Body)
			if err != nil {
				log.Printf("Error parsing message: %v", err.Error())
				continue
			}

			switch msg.Type {
			case "TASK":
				// task handler
			default:
				log.Printf("Unknown msg type: %v", msg.Type)
			}
		}
	}()

	log.Println("Waiting for msgs")

	<-forever
}
