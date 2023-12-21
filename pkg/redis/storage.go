package redis

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Task struct {
	ID            uuid.UUID `json:"id"`
	Status        string    `json:"status"`
	CpuLoad       int       `json:"cpu_load"`
	DiskLoad      int       `json:"disk_load"`
	MemoryLoad    int       `json:"memory_load"`
	BandwidthLoad int       `json:"bandwidth_load"`
	ExecutionTime int       `json:"execution_time"`
}

func (c *RedisClient) StoreTask(task *Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	_, err = c.Set(ctx, task.ID.String(), taskJSON, 0).Result()

	return err
}

func (c *RedisClient) RetrieveTaskByID(taskID uuid.UUID) (*Task, error) {
	taskJSON, err := c.Get(ctx, taskID.String()).Result()
	if err != nil {
		return nil, err
	}

	var task Task

	err = json.Unmarshal([]byte(taskJSON), &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *RedisClient) UpdateTaskStatus(taskID uuid.UUID, status string) error {
	task, err := c.RetrieveTaskByID(taskID)
	if err != nil {
		return err
	}

	task.Status = status

	err = c.StoreTask(task)
	if err != nil {
		return err
	}

	return nil
}
