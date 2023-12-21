package rabbitmq

import (
	"distributed-task-scheduler/pkg/scheduler"
	"distributed-task-scheduler/pkg/server"

	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queue      string
	exchange   string
	bindingKey string
}

func NewConsumer(amqpURL string, queueName string, exchange string, bindingKey string) (*Consumer, error) {
	connection, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
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
		return nil, fmt.Errorf("unable to declare exchange: %w", err)
	}

	q, err := channel.QueueDeclare(
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

	err = channel.QueueBind(
		q.Name,
		bindingKey,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to bind queue: %w", err)
	}

	return &Consumer{
		connection: connection,
		channel:    channel,
		queue:      q.Name,
		exchange:   exchange,
		bindingKey: bindingKey,
	}, nil
}

func (c *Consumer) Consume(redisClient *redis.Client, serverID uuid.UUID) {
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
		log.Fatalf("failed to create a channel: %v", err.Error())
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			msg, err := ParseMessage(d.Body)
			if err != nil {
				log.Printf("Error parsing message: %v", err.Error())
				d.Nack(false, true)
				continue
			}

			switch msg.Type {
			case "TASK_REG":
				message := scheduler.Message{
					ID:   msg.ID,
					Type: msg.Type,
				}
				scheduler.HandleTaskRegistry(message)
			case "TASK_EXECUTE":
				message := server.Message{
					ID:   msg.ID,
					Type: msg.Type,
				}
				server.HandleTaskExecution(message, redisClient, serverID)
			default:
				log.Printf("Unknown msg type: %v", msg.Type)
			}

			d.Ack(false)
		}
	}()

	log.Println("Waiting for msgs")

	<-forever
}
