# Start from a base image with Go installed
FROM golang:1.20.5 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the scheduler app
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler ./cmd/scheduler/main.go

# Build the task generator app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o task-generator ./cmd/task-generator/main.go

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
# COPY --from=builder /app/scheduler .
COPY --from=builder /app/task-generator .

# Copy the .env file into the final image
COPY --from=builder /app/.env .

# Expose ports (if needed)
EXPOSE 8000 8001

# At runtime, select which binary to run
CMD ["./task-generator"]
