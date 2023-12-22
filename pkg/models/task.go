package models

import "github.com/google/uuid"

type Task struct {
	ID            uuid.UUID `json:"id"`
	Status        string    `json:"status"`
	CpuLoad       int       `json:"cpu_load"`
	DiskLoad      int       `json:"disk_load"`
	MemoryLoad    int       `json:"memory_load"`
	ExecutionTime int       `json:"execution_time"`
}
