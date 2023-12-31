package generator

import (
	"distributed-task-scheduler/pkg/models"
	"distributed-task-scheduler/pkg/rabbitmq"
	"distributed-task-scheduler/pkg/redis"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Generator struct {
	taskRegistryProducer *rabbitmq.Producer
	redisClient          *redis.RedisClient
	generateInterval     time.Duration
}

func NewGenerator(amqpURL string, taskRegistryExchange string, interval time.Duration) (*Generator, error) {
	taskRegistryProducer, err := rabbitmq.NewProducer(amqpURL, taskRegistryExchange)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	redisClient := redis.NewClient()

	return &Generator{
		taskRegistryProducer: taskRegistryProducer,
		redisClient:          redisClient,
		generateInterval:     interval,
	}, nil
}

func (g *Generator) generateTask() (*models.Task, error) {
	cpuLoad, err1 := strconv.Atoi(os.Getenv("TASK_CPU_LOAD"))
	diskLoad, err2 := strconv.Atoi(os.Getenv("TASK_DISK_LOAD"))
	memLoad, err3 := strconv.Atoi(os.Getenv("TASK_MEMORY_LOAD"))
	execTime, err4 := strconv.Atoi(os.Getenv("TASK_EXECUTION_TIME"))

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return nil, errors.New("failed to get task metrics")
	}

	return &models.Task{
		ID:            uuid.New(),
		Status:        "pending",
		CpuLoad:       cpuLoad,
		DiskLoad:      diskLoad,
		MemoryLoad:    memLoad,
		ExecutionTime: execTime,
	}, nil
}

func (g *Generator) publishTaskResgistryMessage(taskReg *models.TaskRegistryMessage) error {
	taskRegistryJSON, err := json.Marshal(taskReg)
	if err != nil {
		return fmt.Errorf("error marshalling task: %w", err)
	}

	return g.taskRegistryProducer.PublishMessage(string(taskRegistryJSON), "task_reg")
}

func (g *Generator) saveTaskToRedis(task *models.Task) error {
	err := g.redisClient.StoreTask(task)
	if err != nil {
		return fmt.Errorf("error saving task to redis: %w", err)
	}

	return nil
}

func (g *Generator) Start(runDuration int) {
	ticker := time.NewTicker(g.generateInterval)
	defer ticker.Stop()

	stopTime := time.After(time.Duration(runDuration) * time.Second)

	for {
		select {
		case <-ticker.C:
			task, err := g.generateTask()
			if err != nil {
				fmt.Printf("Error generating task: %v\n", err)
				continue
			}

			err = g.saveTaskToRedis(task)
			if err != nil {
				fmt.Printf("Error saving task to Redis: %v\n", err)
				continue
			}

			taskRegMsgVal := &models.TaskRegistryMessageValue{
				TaskID: task.ID,
			}

			taskRegistry := &models.TaskRegistryMessage{
				ID:    uuid.New(),
				Type:  "TASK_REG",
				Value: *taskRegMsgVal,
			}

			err = g.publishTaskResgistryMessage(taskRegistry)

			if err != nil {
				fmt.Printf("Error publishing task: %s", err.Error())
			}

		case <-stopTime:
			fmt.Println("Simulation Complete")
			return
		}
	}
}
