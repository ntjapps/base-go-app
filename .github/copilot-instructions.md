<!--
  Model instructions for AI agents in base-go-app.
  Keep it short and actionable for quick contributions.
-->

# AI assistant instructions for base-go-app (Go Worker)

Purpose: Help an AI agent be productive and safe in a project with a Go worker processing RabbitMQ tasks. Emphasize strong typing, error handling, and clean architecture.

Project snapshot
- Entry point: `cmd/worker/main.go`
- Configuration: `internal/config` (loads from env)
- Database: `internal/database` (GORM + Postgres)
- Queue: `internal/queue` (RabbitMQ consumer)
- Tasks: `internal/tasks` (Task handlers like `logger_task.go`)
- Models: `internal/models` (GORM models)

Quick workflows (common commands)
- Run worker locally: `go run cmd/worker/main.go`
- Run tests: `go test ./...`
- Tidy modules: `go mod tidy`

Key patterns & constraints
- **Error Handling**: Always check and handle errors. Do not ignore them. Use `log.Printf` or `log.Fatalf` appropriately.
- **Configuration**: Use `internal/config` to access environment variables. Do not use `os.Getenv` directly in business logic.
- **Database**: Use `internal/database.DB` for database operations. Ensure models are defined in `internal/models`.
- **Queue**: The worker uses `amqp091-go`. Ensure consumers handle connection drops or errors gracefully (though basic implementation is provided).
- **JSON Handling**: Be robust with JSON parsing. The worker handles both Celery-style `[[args], kwargs, embed]` and raw JSON payloads.

Typical idioms & references
- **Task Handler**: Create a new function in `internal/tasks` and call it from `internal/queue/consumer.go`.
- **Logging**: Use standard `log` package.
- **UUID**: Use `github.com/google/uuid` for ID generation (V7 preferred).

Tests & CI rules
- Add tests in `_test.go` files next to the code they test.
- Ensure `go test ./...` passes.

Security & safety
- Avoid storing secrets in code. Use environment variables.
- Validate inputs in task handlers.

Files to inspect first
- `cmd/worker/main.go`, `internal/config/config.go`, `internal/queue/consumer.go`, `internal/tasks/logger_task.go`.

If in doubt
- Check `base-python-app` for logic reference if porting features, but adapt to Go idioms.
