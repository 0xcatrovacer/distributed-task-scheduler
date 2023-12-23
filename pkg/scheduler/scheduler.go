package scheduler

import (
	"context"
	"encoding/json"
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

	serverID, err := s.chooseServer()
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

func (s *Scheduler) chooseServer() (string, error) {
	// TODO: Implement Server selection logic

	return uuid.New().String(), nil
}
