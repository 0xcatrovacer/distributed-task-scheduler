package common

import (
	"distributed-task-scheduler/pkg/models"
)

type MessageHandler interface {
	HandleMsg(msg *models.Message)
}

type MessagePublisher interface {
	PublishMessage(message string, routingKey string) error
}

type TaskScheduler interface {
	ScheduleTask(msg *models.Message) error
	UpdateTaskStatus(msg *models.Message) error
}

type TaskExecutor interface {
	HandleTaskExecution(msg *models.Message) error
}
