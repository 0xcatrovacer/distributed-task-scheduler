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

type Task struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	CpuLoad       int    `json:"cpu_load"`
	DiskLoad      int    `json:"disk_load"`
	MemoryLoad    int    `json:"memory_load"`
	BandwidthLoad int    `json:"bandwidth_load"`
	ExecutionTime int    `json:"execution_time"`
}

type TaskRegistry struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type Generator struct {
	taskRegistryProducer *rabbitmq.Producer
	generateInterval     time.Duration
}

func NewGenerator(amqpURL string, taskQueue string, taskRegistryQueue string, interval time.Duration) (*Generator, error) {
	taskRegistryProducer, err := rabbitmq.NewProducer(amqpURL, taskRegistryQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Generator{
		taskRegistryProducer: taskRegistryProducer,
		generateInterval:     interval,
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

	return &Task{
		ID:            uuid.New().String(),
		Status:        "pending",
		CpuLoad:       cpuLoad,
		DiskLoad:      diskLoad,
		MemoryLoad:    memLoad,
		BandwidthLoad: bwLoad,
		ExecutionTime: execTime,
	}, nil
}

func (g *Generator) publishTaskResgistryMessage(taskReg *TaskRegistry) error {
	taskRegistryJSON, err := json.Marshal(taskReg)
	if err != nil {
		return fmt.Errorf("error marshalling task: %w", err)
	}

	return g.taskRegistryProducer.PublishMessage(string(taskRegistryJSON))
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

		taskRegistry := &TaskRegistry{
			ID:   task.ID,
			Type: "TASK_REG",
		}

		err = g.publishTaskResgistryMessage(taskRegistry)

		if err != nil {
			fmt.Printf("Error publishing task")
		}

		fmt.Printf("Published task %v to queues task queue and task registry queue", task.ID)
	}
}
