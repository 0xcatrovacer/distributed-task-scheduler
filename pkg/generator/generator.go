package generator

import (
	"distributed-task-scheduler/pkg/rabbitmq"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
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
	cpuLoad, err1 := strconv.Atoi(os.Getenv("TASK_CPU_LOAD"))
	diskLoad, err2 := strconv.Atoi(os.Getenv("TASK_DISK_LOAD"))
	memLoad, err3 := strconv.Atoi(os.Getenv("TASK_MEMORY_LOAD"))
	bwLoad, err4 := strconv.Atoi(os.Getenv("TASK_BANDWIDTH_LOAD"))
	execTime, err5 := strconv.Atoi(os.Getenv("TASK_EXECUTION_TIME"))

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return nil, errors.New("failed to get task metrics")
	}

	taskMetrics := &TaskMetricsContent{
		CpuLoad:       cpuLoad,
		DiskLoad:      diskLoad,
		MemoryLoad:    memLoad,
		BandwidthLoad: bwLoad,
		ExecutionTime: execTime,
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