package main

import (
	"distributed-task-scheduler/pkg/rabbitmq"
	"distributed-task-scheduler/pkg/server"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type ServerStatus struct {
	ID                uuid.UUID `json:"id"`
	CpuUtilization    int       `json:"cpu_utilization"`
	CpuLimit          int       `json:"cpu_limit"`
	MemoryUtilization int       `json:"mem_utilization"`
	MemoryLimit       int       `json:"mem_limit"`
	DiskUtilization   int       `json:"disk_utilization"`
	DiskLimit         int       `json:"disk_limit"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error reading env variables: %v", err)
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
	producer, err := rabbitmq.NewProducer(amqpURL, "completed_exchange")
	if err != nil {
		log.Fatalf("Could not create rabbitmq producer: %v", err.Error())
	}

	serverID := uuid.New()
	queueName := "schedule_queue." + serverID.String()

	server := server.New(redisClient, producer, serverID)

	err = server.UpdateInitialComputeInfo()
	if err != nil {
		log.Fatalf("Error updating initial compute info: %v", err.Error())
	}

	consumer, err := rabbitmq.NewConsumer(amqpURL, queueName, "schedule_exchange", serverID.String())
	if err != nil {
		log.Fatalf("Error creating new consumer: %v", err.Error())
	}

	consumer.Consume(nil, server)
}
