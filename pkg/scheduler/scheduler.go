package scheduler

import "github.com/google/uuid"

type Message struct {
	ID   uuid.UUID
	Type string
}

func HandleTaskRegistry(msg Message) {
	//TODO: To be implemented
}
