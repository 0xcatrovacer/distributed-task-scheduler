package main

import (
	"distributed-task-scheduler/pkg/generator"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err.Error())
	}

	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	taskQueue := "Task_Queue"
	taskRegistryQueue := "Task_Registry_Queue"

	generationInterval, err := strconv.Atoi(os.Getenv("TASK_GENERATION_INTERVAL"))
	if err != nil {
		log.Fatalf("Error getting task generation interval: %v", err.Error())
	}

	generateInterval := time.Duration(generationInterval) * time.Millisecond

	generator, err := generator.NewGenerator(amqpURL, taskQueue, taskRegistryQueue, generateInterval)
	if err != nil {
		log.Fatalf("Failed to create task generator: %v", err)
	}

	generator.Start()
}
