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

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
