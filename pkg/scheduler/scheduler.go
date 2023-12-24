package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"distributed-task-scheduler/pkg/models"
	"distributed-task-scheduler/pkg/rabbitmq"
)

type Scheduler struct {
	redisClient *redis.Client
	producer    *rabbitmq.Producer
}

func New(redisClient *redis.Client, producer *rabbitmq.Producer) *Scheduler {
	return &Scheduler{
		redisClient: redisClient,
		producer:    producer,
	}
}

func (s *Scheduler) ScheduleTask(msg *models.Message) error {
	var value models.TaskRegistryMessageValue

	if err := json.Unmarshal(msg.Value, &value); err != nil {
		return fmt.Errorf("error unmarshalling task registry message: %w", err)
	}

	taskRegistryMsg := models.TaskRegistryMessage{
		ID:    msg.ID,
		Type:  msg.Type,
		Value: value,
	}

	task, err := s.fetchTask(taskRegistryMsg.Value.TaskID.String())
	if err != nil {
		return fmt.Errorf("error fetching task from Redis: %w", err)
	}

	serverID, err := s.chooseServer(task)
	if err != nil {
		return fmt.Errorf("error choosing server: %w", err)
	}

	scheduleMsgVal := models.ScheduleMessageValue{
		TaskID:   taskRegistryMsg.Value.TaskID,
		ServerID: uuid.MustParse(serverID),
	}

	scheduleMsg := models.ScheduleMessage{
		ID:    uuid.New(),
		Type:  "TASK_EXECUTE",
		Value: scheduleMsgVal,
	}

	msgBytes, err := json.Marshal(scheduleMsg)
	if err != nil {
		return fmt.Errorf("error marshaling schedule message: %w", err)
	}

	routingKey := serverID
	if err := s.producer.PublishMessage(string(msgBytes), routingKey); err != nil {
		return fmt.Errorf("error publishing schedule message: %w", err)
	}

	log.Printf("Scheduled task %s to server %s", task.ID, serverID)

	return nil
}

func (s *Scheduler) fetchTask(taskID string) (*models.Task, error) {
	taskData, err := s.redisClient.Get(context.Background(), "task:"+taskID).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving task from Redis: %w", err)
	}

	var task models.Task
	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		return nil, fmt.Errorf("error unmarshaling task data: %w", err)
	}
	return &task, nil
}

func (s *Scheduler) chooseServer(task *models.Task) (string, error) {
	serverKeys, err := s.redisClient.Keys(context.Background(), "server:*").Result()
	if err != nil {
		return "", fmt.Errorf("error getting keys: %w", err)
	}

	chosenServer := &models.ServerStatus{}
	minCpuRatio := float64(1)
	minDiskRatio := float64(1)

	for _, key := range serverKeys {
		var server models.ServerStatus

		serverData, err := s.redisClient.Get(context.Background(), key).Result()
		if err != nil {
			return "", fmt.Errorf("error retrieving value for key: %s", err)
		}

		if err = json.Unmarshal([]byte(serverData), &server); err != nil {
			return "", fmt.Errorf("error unmarshaling server data: %w", err)
		}

		cpuUtil := server.CpuUtilization
		cpuLimit := server.CpuLimit
		memUtil := server.MemoryUtilization
		memLimit := server.MemoryLimit
		diskUtil := server.DiskUtilization
		diskLimit := server.DiskLimit

		newCpuRatio := float64((cpuUtil + task.CpuLoad) / cpuLimit)
		newMemRatio := float64((memUtil + task.MemoryLoad) / memLimit)
		newDiskRatio := float64((diskUtil + task.DiskLoad) / diskLimit)

		if newCpuRatio < 0.9 && newMemRatio < 0.9 && newDiskRatio <= 0.9 {
			if newCpuRatio < minCpuRatio {
				minCpuRatio = newCpuRatio
				minDiskRatio = newDiskRatio
				chosenServer = &server
			} else if newCpuRatio == minCpuRatio && newDiskRatio < minDiskRatio {
				minCpuRatio = newCpuRatio
				minDiskRatio = newDiskRatio
				chosenServer = &server
			}
		}
	}

	if chosenServer.ID == uuid.Nil {
		return "", errors.New("could not find free server")
	}

	return chosenServer.ID.String(), nil
}

func (s *Scheduler) UpdateTaskStatus(msg *models.Message) error {
	var taskCompleteMsgValue models.UpdateMessageValue

	if err := json.Unmarshal([]byte(msg.Value), &taskCompleteMsgValue); err != nil {
		return fmt.Errorf("error in unmarshalling msg value: %w", err)
	}

	var task models.Task

	taskData, err := s.redisClient.Get(context.Background(), "task:"+taskCompleteMsgValue.TaskID.String()).Result()
	if err != nil {
		return fmt.Errorf("error in fetching task from redis: %w", err)
	}

	if err := json.Unmarshal([]byte(taskData), &task); err != nil {
		return fmt.Errorf("error in unmarshalling task data: %w", err)
	}

	task.Status = taskCompleteMsgValue.Status

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error in marshalling task data: %w", err)
	}

	err = s.redisClient.Set(context.Background(), "task:"+taskCompleteMsgValue.TaskID.String(), data, 0).Err()
	if err != nil {
		return fmt.Errorf("error in setting task data to redis: %w", err)
	}

	fmt.Println("Updated task " + task.ID.String() + " to status " + taskCompleteMsgValue.Status)

	return nil
}
