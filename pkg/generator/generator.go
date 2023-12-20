package generator

import (
	"distributed-task-scheduler/pkg/rabbitmq"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskMetricsContent struct {
	CpuLoad       int `json:"cpu_load"`
	DiskLoad      int `json:"disk_load"`
	MemoryLoad    int `json:"memory_load"`
	BandwidthLoad int `json:"bandwidth_load"`
	ExecutionTime int `json:"execution_time"`
}

type Task struct {
	ID      string             `json:"id"`
	Type    string             `json:"type"`
	Metrics TaskMetricsContent `json:"metrics"`
}

type Generator struct {
	producer         *rabbitmq.Producer
	generateInterval time.Duration
}

func NewGenerator(amqpURL string, queueName string, interval time.Duration) (*Generator, error) {
	producer, err := rabbitmq.NewProducer(amqpURL, queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Generator{
		producer:         producer,
		generateInterval: interval,
	}, nil
}

func (g *Generator) generateTask() (*Task, error) {
	taskMetrics := &TaskMetricsContent{
		CpuLoad:       1,
		DiskLoad:      1,
		MemoryLoad:    1,
		BandwidthLoad: 1,
		ExecutionTime: 10,
	}

	return &Task{
		ID:      uuid.New().String(),
		Type:    "TASK",
		Metrics: *taskMetrics,
	}, nil
}

func (g *Generator) publishTask(task *Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error marshalling task: %w", err)
	}

	return g.producer.PublishMessage(string(taskJSON))
}

func (g *Generator) Start() {
	ticker := time.NewTicker(g.generateInterval)
	defer ticker.Stop()

	for range ticker.C {
		task, err := g.generateTask()
		if err != nil {
			fmt.Printf("Error generating task: %v\n", err)
			continue
		}

		err = g.publishTask(task)
		if err != nil {
			fmt.Printf("Error publishing task: %v\n", err)
		}

		fmt.Printf("Published task %v to queue", task.ID)
	}
}
