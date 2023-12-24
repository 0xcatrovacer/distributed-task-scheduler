# Distributed Task Scheduler Simulator

## About

This is a distributed task scheduler simulator written in Golang. It uses Golang's concurrency model to simulate scheduling tasks to different servers.

![Design](https://miscellanous-public.s3.ap-south-1.amazonaws.com/images/dtss/distributed-task-scheduler-final.png)

### Components:

**`Task Generator`:** Generates tasks and sends them to Redis to be scheduled. It also sends task IDs to a `registry queue`. 

**`Scheduler`:** Consumes messages from `registry queue`. For every task, gets compute availability info of all servers from Redis. Chooses which server is the most available, and sends task ID and server ID to `schedule exchange` with the server id as the routing key so that servers can fetch tasks assigned to them. <br> Also consumes messages from `completed queue` and updates task status in redis.

**`Server`:** Fetches task IDs from `schedule queue` and fetches the task from Redis using the task ID. After executing the task, the server publishes the task ID to `completed queue`

## Getting Started

### Prerequisites

Make sure you have Docker installed

### Installation

Clone the repo:

```bash
git clone https://github.com/0xcatrovacer/distributed-task-scheduler
```

Create a .env file and set the environment variables

```bash
touch .env
```

```.env
# SIMULATION CONFIGURATIONS
SIMULATION_RUN_DURATION=           # (in s) Duration for which simulation will run 

# RABBITMQ CONFIGURATIONS
RABBITMQ_HOST=                     # RabbitMQ host address 
RABBITMQ_PORT=                     # RabbitMQ port
RABBITMQ_DEFAULT_USER=             # RabbitMQ User
RABBITMQ_DEFAULT_PASS=             # RabbitMQ Password

# REDIS CONFIGURATIONS
REDIS_ADDRESS=                     # Redis Address (addr:port) 
REDIS_PASSWORD=                    # Redis Password 

# TASK CONFIGURATIONS
TASK_CPU_LOAD=                      # (in %) Load 1 task will put on server's CPU
TASK_DISK_LOAD=                     # (in MBs) Load 1 task will put on server's disk
TASK_MEMORY_LOAD=                   # (in MBs) Load 1 task will put on server's memory
TASK_EXECUTION_TIME=                # (in ms) Time for which a task will run

TASK_GENERATION_INTERVAL=

# SERVER CONFIGURATIONS
INITIAL_CPU_UTILIZATION=            # (in %) Initial load on a server's CPU
CPU_LIMIT=                          # (in %) Server's CPU limit. Should not be more than 100
INITIAL_MEMORY_UTILIZATION=         # (in MBs) Initial load on a server's memory
MEMORY_LIMIT=                       # (in MBs) Server's memory limit
INITIAL_DISK_UTILIZATION=           # (in MBs) Initial load on a server's disk
DISK_LIMIT=                         # (in MBs) Server's disk limit
```

Run using docker

```bash
docker-compose up --scale server=n 
# n = number of servers
```

## Usage

RabbitMQ details can be checked in the browser at AMQP address

![RabbitMQ](https://miscellanous-public.s3.ap-south-1.amazonaws.com/images/dtss/rabbitmq-ss.png)

Redis details can be accessed with any Redis GUI. I used RedisInsight.

![Redis](https://miscellanous-public.s3.ap-south-1.amazonaws.com/images/dtss/redis-ss.png)
