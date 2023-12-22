package redis

import (
	"distributed-task-scheduler/pkg/models"
	"encoding/json"

	"github.com/google/uuid"
)

func (c *RedisClient) StoreTask(task *models.Task) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	_, err = c.Set(ctx, "task:"+task.ID.String(), taskJSON, 0).Result()

	return err
}

func (c *RedisClient) RetrieveTaskByID(taskID uuid.UUID) (*models.Task, error) {
	taskJSON, err := c.Get(ctx, taskID.String()).Result()
	if err != nil {
		return nil, err
	}

	var task models.Task

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
