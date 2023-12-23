package models

import "github.com/google/uuid"

type ServerStatus struct {
	ID                uuid.UUID `json:"id"`
	CpuUtilization    int       `json:"cpu_utilization"`
	CpuLimit          int       `json:"cpu_limit"`
	MemoryUtilization int       `json:"mem_utilization"`
	MemoryLimit       int       `json:"mem_limit"`
	DiskUtilization   int       `json:"disk_utilization"`
	DiskLimit         int       `json:"disk_limit"`
}
