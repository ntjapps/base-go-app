# Base Go App - Worker

This is a Go implementation of the worker system, replacing the Python Celery worker.
It is optimized for RabbitMQ and currently handles the `log_db_task`.

## Prerequisites

- Go 1.21+
- RabbitMQ
- PostgreSQL
- `.env` file (see `base-go-app/.env.example` or copy from there)

## Structure

- `cmd/worker/main.go`: Entry point.
- `internal/config`: Configuration loading.
- `internal/database`: Database connection.
- `internal/models`: Data models.
- `internal/queue`: RabbitMQ consumer.
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

### `log_db_task`

Inserts a log record into the `log` table.
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
