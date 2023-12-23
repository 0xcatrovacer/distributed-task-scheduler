package main

import (
	"distributed-task-scheduler/pkg/rabbitmq"
	"distributed-task-scheduler/pkg/scheduler"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err.Error())
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	taskRegistryQueue := "task_registry_queue"
	taskExchange := "task_registry_exchange"
	routingKey := "task_reg"

	consumer, err := rabbitmq.NewConsumer(amqpURL, taskRegistryQueue, taskExchange, routingKey)
	if err != nil {
		log.Fatalf("Failed to create consumer: %s", err)
	}

	producer, err := rabbitmq.NewProducer(amqpURL, "schedule_exchange")
	if err != nil {
		log.Fatalf("error creating RabbitMQ producer: %s", err)
	}

	taskScheduler := scheduler.New(redisClient, producer)

	consumer.Consume(taskScheduler, nil)
}
