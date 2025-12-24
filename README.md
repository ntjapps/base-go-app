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

---

## Features ✅

- RabbitMQ-only worker optimized to receive Celery-compatible payloads.
- PostgreSQL persistence using GORM.
- Healthcheck HTTP endpoint (`/healthz`) for container orchestration.
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
- Includes a Docker `HEALTHCHECK` that calls `GET /healthz`.

Build and run locally:

```bash
# Build image
docker build -t myorg/base-go-app:staging .

# Run container
docker run -e RABBITMQ_HOST=... -e DB_HOST=... -p 8080:8080 myorg/base-go-app:staging

# Check health
curl http://localhost:8080/healthz
```

## CI / CD ⚙️

- `.github/workflows/docker-build-staging.yaml` — manual/staging container build (tags `staging`).
- `.github/workflows/docker-build-prod.yaml` — triggered on semver tags and pushes production images to `ghcr.io`.
- `.github/dependabot.yml` — keeps Go modules, Actions and Dockerfiles updated automatically.

---

If you'd like, I can also:
- Add a `readiness` endpoint that checks deeper operational readiness.
- Publish release artifacts (tar/zip) alongside container images.
