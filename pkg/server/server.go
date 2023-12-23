package server

import (
	"context"
	"distributed-task-scheduler/pkg/models"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type ServerStatus struct {
	ID                uuid.UUID `json:"id"`
	CpuUtilization    int       `json:"cpu_utilization"`
	CpuLimit          int       `json:"cpu_limit"`
	MemoryUtilization int       `json:"mem_utilization"`
	MemoryLimit       int       `json:"mem_limit"`
	DiskUtilization   int       `json:"disk_utilization"`
	DiskLimit         int       `json:"disk_limit"`
}

type Server struct {
	redisClient *redis.Client
	serverID    uuid.UUID
}

func New(redisClient *redis.Client, serverID uuid.UUID) *Server {
	return &Server{
		redisClient: redisClient,
		serverID:    serverID,
	}
}

func (s *Server) HandleTaskExecution(msg *models.Message) error {
	// TODO: Implement task fetching and execution logic
	// TODO: Implement new compute info calculation logic

	newCpuUtil := 0
	newMemUtil := 0
	newDiskUtil := 0

	s.UpdateComputeInfo(s.serverID, newCpuUtil, newMemUtil, newDiskUtil)

	return nil
}

func (s *Server) UpdateInitialComputeInfo() {
	cpuUtil, err1 := strconv.Atoi(os.Getenv("INITIAL_CPU_UTILIZATION"))
	cpuLimit, err2 := strconv.Atoi(os.Getenv("CPU_LIMIT"))
	memUtil, err3 := strconv.Atoi(os.Getenv("INITIAL_MEMORY_UTILIZATION"))
	memLimit, err4 := strconv.Atoi(os.Getenv("MEMORY_LIMIT"))
	diskUtil, err5 := strconv.Atoi(os.Getenv("INITIAL_DISK_UTILIZATION"))
	diskLimit, err6 := strconv.Atoi(os.Getenv("DISK_LIMIT"))

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		log.Println("unable to fetch initial compute info")
	}

	computeInfo := ServerStatus{
		ID:                s.serverID,
		CpuUtilization:    cpuUtil,
		CpuLimit:          cpuLimit,
		MemoryUtilization: memUtil,
		MemoryLimit:       memLimit,
		DiskUtilization:   diskUtil,
		DiskLimit:         diskLimit,
	}

	data, err := json.Marshal(computeInfo)
	if err != nil {
		log.Printf("Error marshalling compute info: %s", err.Error())
		return
	}

	err = s.redisClient.Set(context.Background(), "server:"+s.serverID.String(), data, 0).Err()
	if err != nil {
		log.Printf("Error in updating compute info to redis: %s", err.Error())
	}
}

func (s *Server) UpdateComputeInfo(serverID uuid.UUID, newCpuUtil int, newMemUtil int, newDiskUtil int) {
	var computeInfo *ServerStatus

	redata, err := s.redisClient.Get(context.Background(), "server:"+serverID.String()).Result()
	if err != nil {
		log.Printf("compute info not present for server: %s", err.Error())
	}

	err = json.Unmarshal([]byte(redata), computeInfo)
	if err != nil {
		log.Printf("could not bind data: %s", err.Error())
	}

	computeInfo.CpuUtilization = newCpuUtil
	computeInfo.MemoryUtilization = newMemUtil
	computeInfo.DiskUtilization = newDiskUtil

	data, err := json.Marshal(computeInfo)
	if err != nil {
		log.Printf("Error marshalling compute info: %s", err.Error())
		return
	}

	err = s.redisClient.Set(context.Background(), "server:"+serverID.String(), data, 0).Err()
	if err != nil {
		log.Printf("Error in updating compute info to redis: %s", err.Error())
	}
}
