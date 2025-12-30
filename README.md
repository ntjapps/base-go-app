# Base Go App - Worker

This is a Go implementation of the worker system, replacing the Python Celery worker.
It is optimized for RabbitMQ and currently handles the `logger` task (stores logs in the database).

## Prerequisites

- Go 1.25+
- RabbitMQ
- PostgreSQL
- `.env` file (create one with the environment variables below or set them in the environment; the app will also load a `.env` file if present) 

## Structure

- `cmd/worker/main.go`: Entry point.
- `internal/config`: Configuration loading.
- `internal/database`: Database connection.
- `internal/models`: Data models.
- `internal/queue`: RabbitMQ consumer.
- `internal/publisher`: RabbitMQ publisher for sending tasks.
- `internal/tasks`: Task handlers.
- `internal/helpers`: Helper functions.

## Running

```bash
go run cmd/worker/main.go
```

## Configuration

The application uses the following environment variables (compatible with the Python app):

- `RABBITMQ_USER`
- `RABBITMQ_PASSWORD`
- `RABBITMQ_HOST`
- `RABBITMQ_PORT`
- `RABBITMQ_VHOST`
- `DB_USERNAME`
- `DB_PASSWORD`
- `DB_HOST`
- `DB_PORT`
- `DB_DATABASE`

## Queues

The worker listens on the `logger` queue with routing key `logger` on exchange `celery`.
It expects messages to be either:
1. Celery format: `[[payload], {}, null]`
2. Raw JSON payload: `payload`

## Tasks

### `logger` task

Inserts a log record into the `log` table (handler registered as `logger`).
Payload structure:
```json
{
    "message": "...",
    "channel": "...",
    "level": "...",
    "level_name": "...",
    "datetime": "...",
    "context": {},
    "extra": {}
}
```

---

## Publishing Tasks to RabbitMQ 📤

The `internal/publisher` package provides functions to submit tasks to RabbitMQ, similar to the Laravel `CeleryFunction` and `GoWorkerFunction` traits.

### Usage Example

```go
package main

import (
    "log"
    "base-go-app/internal/config"
    "base-go-app/internal/publisher"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Create publisher
    pub, err := publisher.NewPublisher(cfg)
    if err != nil {
        log.Fatalf("Failed to create publisher: %v", err)
    }
    defer pub.Close()
    
    // Send Celery task (for Python workers)
    taskID, err := pub.SendCeleryTask(
        "celery_test_task",
        []interface{}{"arg1", "arg2"},
        "celery",
    )
    if err != nil {
        log.Fatalf("Failed to send Celery task: %v", err)
    }
    log.Printf("Celery task submitted with ID: %s", taskID)
    
    // Send Go worker task
    timeout := 300
    options := &publisher.TaskOptions{
        TimeoutSeconds: &timeout,
        Notify: map[string]string{
            "webhook": "http://example.com/callback",
        },
    }
    
    payload := map[string]interface{}{
        "message":    "test log",
        "level":      "200",
        "level_name": "INFO",
    }
    
    taskID, err = pub.SendGoTask("logger", payload, "logger", options)
    if err != nil {
        log.Fatalf("Failed to send Go task: %v", err)
    }
    log.Printf("Go task submitted with ID: %s", taskID)
}
```

### Publisher Functions

#### `SendCeleryTask(task, args, queue) (taskID, error)`
Sends a task in Celery protocol v2 format for Python workers. Compatible with the Laravel `CeleryFunction::sendTask()`.

**Parameters:**
- `task`: Task name (e.g., "celery_test_task")
- `args`: Array of arguments for the task
- `queue`: RabbitMQ queue name (default: "celery")

**Returns:** Task ID (UUID) and error if any

#### `SendGoTask(task, payload, queue, options) (taskID, error)`
Sends a task in Go worker format. Compatible with the Laravel `GoWorkerFunction::sendGoTask()`.

**Parameters:**
- `task`: Task name (e.g., "logger")
- `payload`: Map of task payload data
- `queue`: RabbitMQ queue name (default: "celery")
- `options`: Optional task options (timeout, notify, max_attempts)

**Returns:** Task ID (UUID) and error if any

### Multiple Queue Support

Both functions support sending tasks to **any queue** for parallel processing. Different task types can be routed to different queues with dedicated workers:

```go
// Logger tasks → go.logger queue
pub.SendGoTask("logger", logPayload, "go.logger", nil)

// Data updates → go.data_processor queue
pub.SendGoTask("update_data", dataPayload, "go.data_processor", nil)

// Emails → go.email queue
pub.SendGoTask("send_email", emailPayload, "go.email", nil)

// Python tasks → python.analytics queue
pub.SendCeleryTask("generate_report", args, "python.analytics")
```

📖 See [MULTIPLE_QUEUES.md](MULTIPLE_QUEUES.md) for detailed examples and architecture patterns.

### Multi-Pod/Container Support

The worker is **production-ready** for running multiple instances (pods/containers) consuming from the same queue. RabbitMQ automatically distributes messages across all consumers with:
- ✅ **No race conditions** - Each message goes to exactly one worker
- ✅ **No duplicate processing** - Manual acknowledgment prevents conflicts
- ✅ **Crash recovery** - Unacknowledged messages are automatically redelivered
- ✅ **Auto-scaling** - Scale from 1 to 100+ pods safely

```bash
# Scale horizontally
kubectl scale deployment go-worker --replicas=20

# Or with Docker Compose
docker-compose up -d --scale go-worker=10
```

📖 See [MULTI_POD_DEPLOYMENT.md](MULTI_POD_DEPLOYMENT.md) for Kubernetes, Docker Swarm, and scaling strategies.

---

## Features ✅

- RabbitMQ-only worker optimized to receive Celery-compatible payloads.
- PostgreSQL persistence using GORM.
- **Multi-pod/container safe** - Scale to 100+ instances without conflicts.
- **Multiple queue support** - Parallel processing across different queues.
- Healthcheck HTTP endpoint (`/healthcheck`) for container orchestration.
- Docker multi-stage build producing a minimal runtime image.
- GitHub Actions for build/test and container publishing.
- Dependabot config to keep Go modules, GitHub Actions and Docker up-to-date.

## Endpoints 🔧

- GET /healthcheck
  - Returns JSON with quick status of the service, database connectivity and RabbitMQ consumer connection state.
  - Example response:

```json
{
  "status": "ok",
  "database": true,
  "rabbitmq": true
}
```

If either the DB ping or RabbitMQ connection check fails, `/healthcheck` will return a non-200 status and include `database_error` when relevant.

## Docker image & Healthcheck 🐳

A multi-stage `Dockerfile` builds a statically-linked Go binary and produces a small Alpine-based image.

- Exposes port `8080` (configurable via `HEALTH_PORT` env var).
- Includes a Docker `HEALTHCHECK` that calls `GET /healthcheck`.

Build and run locally:

```bash
# Build image
docker build -t myorg/base-go-app:staging .

# Run container
docker run -e RABBITMQ_HOST=... -e DB_HOST=... -p 8080:8080 myorg/base-go-app:staging

# Check health
curl http://localhost:8080/healthcheck
```

## CI / CD ⚙️

- `.github/workflows/docker-build-staging.yaml` — manual/staging container build (tags `staging`).
- `.github/workflows/docker-build-prod.yaml` — triggered on semver tags and pushes production images to `ghcr.io`.
- `.github/dependabot.yml` — keeps Go modules, Actions and Dockerfiles updated automatically.

---

If you'd like, I can also:
- Add a `readiness` endpoint that checks deeper operational readiness.
- Publish release artifacts (tar/zip) alongside container images.
