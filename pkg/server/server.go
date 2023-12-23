package server

import (
	"context"
	"distributed-task-scheduler/pkg/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

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
	var value models.ScheduleMessageValue

	if err := json.Unmarshal(msg.Value, &value); err != nil {
		return fmt.Errorf("error unmarshalling schedule message: %w", err)
	}

	scheduleMsg := models.ScheduleMessage{
		ID:    msg.ID,
		Type:  msg.Type,
		Value: value,
	}

	serverID := scheduleMsg.Value.ServerID
	if serverID != s.serverID {
		return fmt.Errorf("task not assigned to this server")
	}

	serverStat, err := s.fetchServer(serverID.String())
	if err != nil {
		return fmt.Errorf("could not fetch server")
	}

	cpuUtil := serverStat.CpuUtilization
	diskUtil := serverStat.DiskUtilization
	memUtil := serverStat.MemoryUtilization

	taskID := scheduleMsg.Value.TaskID

	task, err := s.fetchTask(taskID.String())
	if err != nil {
		return fmt.Errorf("could not fetch task")
	}

	newCpuUtil := cpuUtil + task.CpuLoad
	newMemUtil := memUtil + task.MemoryLoad
	newDiskUtil := diskUtil + task.DiskLoad

	s.UpdateComputeInfo(s.serverID, newCpuUtil, newMemUtil, newDiskUtil)

	execution := time.After(time.Duration(task.ExecutionTime))

	<-execution

	newCpuUtil -= task.CpuLoad
	newMemUtil -= task.MemoryLoad

	s.UpdateComputeInfo(s.serverID, newCpuUtil, newMemUtil, newDiskUtil)

	// TODO: Add logic to send to completed queue

	return nil
}

func (s *Server) fetchTask(taskID string) (*models.Task, error) {
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

func (s *Server) fetchServer(serverID string) (*models.ServerStatus, error) {
	serverData, err := s.redisClient.Get(context.Background(), "server:"+serverID).Result()
	if err != nil {
		return nil, fmt.Errorf("error retrieving server from Redis: %w", err)
	}

	var serverStat models.ServerStatus
	if err := json.Unmarshal([]byte(serverData), &serverStat); err != nil {
		return nil, fmt.Errorf("error unmarshaling task data: %w", err)
	}
	return &serverStat, nil
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

	computeInfo := models.ServerStatus{
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
	computeInfo := &models.ServerStatus{}

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
