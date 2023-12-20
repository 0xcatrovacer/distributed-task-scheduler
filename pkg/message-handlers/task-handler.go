package messagehandlers

type TaskMessageMetrics struct {
	CpuLoad       int `json:"cpu_load"`
	DiskLoad      int `json:"disk_load"`
	MemoryLoad    int `json:"memory_load"`
	BandwidthLoad int `json:"bandwidth_load"`
	ExecutionTime int `json:"execution_time"`
}

func (h *TaskMessageMetrics) HandleMessage() {
	// To be implemented
}
