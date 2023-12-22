package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Message struct {
	ID    uuid.UUID
	Type  string
	Value json.RawMessage `json:"value"`
}

type TaskRegistryMessageValue struct {
	TaskID uuid.UUID `json:"task_id"`
}

type TaskRegistryMessage struct {
	ID    uuid.UUID                `json:"id"`
	Type  string                   `json:"type"`
	Value TaskRegistryMessageValue `json:"value"`
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

type ScheduleMessageValue struct {
	TaskID   uuid.UUID `json:"task_id"`
	ServerID uuid.UUID `json:"server_id"`
}

type ScheduleMessage struct {
	ID    uuid.UUID            `json:"id"`
	Type  string               `json:"type"`
	Value ScheduleMessageValue `json:"value"`
}
