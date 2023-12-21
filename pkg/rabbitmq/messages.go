package rabbitmq

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Message struct {
	ID      uuid.UUID
	Type    string
	Metrics json.RawMessage
}

type RegistryMessage struct {
	ID   uuid.UUID
	Type string
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func ParseRegistryMessage(data []byte) (*RegistryMessage, error) {
	var msg RegistryMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
