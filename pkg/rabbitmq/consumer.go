package rabbitmq

import (
	"distributed-task-scheduler/pkg/common"
	"distributed-task-scheduler/pkg/models"

	"fmt"
	"log"

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

func (c *Consumer) Consume(scheduler common.TaskScheduler, executor common.TaskExecutor) {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
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
			msg, err := models.ParseMessage(d.Body)
			if err != nil {
				log.Printf("Error parsing message: %v", err.Error())
				d.Nack(false, true)
				continue
			}

			switch msg.Type {
			case "TASK_REG":
				if err = scheduler.ScheduleTask(msg); err != nil {
					log.Printf("Error scheduling task: %v", err)
					d.Nack(false, true)
					continue
				}
			case "TASK_EXECUTE":
				if err = executor.HandleTaskExecution(msg); err != nil {
					log.Printf("Error executing task: %v", err)
					d.Nack(false, true)
					continue
				}
			case "TASK_COMPLETED":
				if err = scheduler.UpdateTaskStatus(msg); err != nil {
					log.Printf("Error executing task: %v", err)
					d.Nack(false, true)
					continue
				}
			default:
				log.Printf("Unknown msg type: %v", msg.Type)
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}()

	log.Println("Waiting for msgs")

	<-forever
}
