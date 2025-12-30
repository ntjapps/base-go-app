# RabbitMQ Publisher Functions Added to base-go-app

## Summary

Added RabbitMQ task publishing functions to base-go-app, matching the functionality of Laravel's `CeleryFunction` and `GoWorkerFunction` traits.

## What Was Added

### 1. Publisher Package (`internal/publisher/`)

#### Files Created:
- **interface.go** - Publisher interface definition
- **publisher.go** - Implementation with `SendCeleryTask()` and `SendGoTask()` functions
- **publisher_test.go** - Comprehensive unit and integration tests
- **example_usage.go** - Inline documentation examples

### 2. Example Application (`cmd/publisher-example/`)

- **main.go** - Standalone example demonstrating all publisher features

### 3. Documentation

- **PUBLISHER_COMPARISON.md** - Side-by-side comparison of PHP vs Go implementations
- **README.md** - Updated with publisher usage section

## Functions Available

### `SendCeleryTask(task, args, queue) (taskID, error)`
Sends tasks to Python Celery workers using Celery protocol v2.

**Compatible with:** Laravel `CeleryFunction::sendTask()`

```go
taskID, err := pub.SendCeleryTask(
    "celery_test_task",
    []interface{}{"arg1", "arg2"},
    "celery",
)
```

### `SendGoTask(task, payload, queue, options) (taskID, error)`
Sends tasks to Go workers using the Go worker format.

**Compatible with:** Laravel `GoWorkerFunction::sendGoTask()`

```go
timeout := 300
options := &publisher.TaskOptions{
    TimeoutSeconds: &timeout,
    Notify: map[string]string{
        "webhook": "http://example.com/callback",
    },
}

taskID, err := pub.SendGoTask(
    "logger",
    map[string]interface{}{
        "message": "test log",
        "level":   "200",
    },
    "logger",
    options,
)
```

## Message Format Compatibility

### ✅ Celery Protocol v2 (for Python workers)
Produces identical message format to Laravel's `CeleryFunction`:
- Body: `[[args...], {}, {metadata}]`
- Headers: `lang`, `task`, `id`, `root_id`
- Properties: `correlation_id`, `content_type`, `content_encoding`

### ✅ Go Worker Format
Produces identical message format to Laravel's `GoWorkerFunction`:
```json
{
    "version": "1.0",
    "id": "uuid",
    "task": "task_name",
    "payload": {...},
    "created_at": "ISO8601",
    "attempt": 0,
    "max_attempts": 5,
    "timeout_seconds": 300,
    "notify": {...}
}
```

## Usage Example

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
        log.Fatal(err)
    }
    
    // Create publisher (reuse for multiple tasks)
    pub, err := publisher.NewPublisher(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer pub.Close()
    
    // Send Celery task
    taskID, err := pub.SendCeleryTask(
        "celery_test_task",
        []interface{}{"arg1", "arg2"},
        "celery",
    )
    
    // Send Go worker task
    taskID, err = pub.SendGoTask(
        "logger",
        map[string]interface{}{
            "message": "log message",
            "level":   "200",
        },
        "logger",
        nil,
    )
}
```

## Testing

All functions include comprehensive tests:

```bash
# Run unit tests (no RabbitMQ required)
go test ./internal/publisher/... -short -v

# Run all tests including integration tests (requires RabbitMQ)
go test ./internal/publisher/... -v
```

Test coverage:
- ✅ Input validation
- ✅ Error handling
- ✅ Message format verification
- ✅ Integration with RabbitMQ
- ✅ Celery protocol compatibility
- ✅ Go worker format compatibility

## Key Features

1. **Full Laravel Compatibility** - Produces identical message formats to Laravel traits
2. **Connection Reuse** - Single publisher can send multiple tasks efficiently
3. **Error Handling** - Returns Go-style errors for robust error handling
4. **Type Safety** - Strongly-typed options and parameters
5. **Well Tested** - Comprehensive unit and integration tests
6. **Documented** - Inline documentation and examples
7. **Production Ready** - Follows Go best practices

## Dependencies Added

- `github.com/google/uuid` v1.6.0 - For UUID generation

All other dependencies were already present in the project.

## Files Modified

1. `/home/GitProjects/base-go-app/README.md` - Added publisher documentation
2. `/home/GitProjects/base-go-app/go.mod` - Added google/uuid dependency

## Files Created

1. `/home/GitProjects/base-go-app/internal/publisher/interface.go`
2. `/home/GitProjects/base-go-app/internal/publisher/publisher.go`
3. `/home/GitProjects/base-go-app/internal/publisher/publisher_test.go`
4. `/home/GitProjects/base-go-app/internal/publisher/example_usage.go`
5. `/home/GitProjects/base-go-app/cmd/publisher-example/main.go`
6. `/home/GitProjects/base-go-app/PUBLISHER_COMPARISON.md`

## Verification

✅ All packages build successfully
✅ All tests pass (unit + integration)
✅ Code follows Go conventions
✅ Compatible with Laravel implementations
✅ Documentation complete

## Next Steps

The publisher is ready to use! You can:

1. Use it in your Go applications to publish tasks to RabbitMQ
2. Run the example: `go run cmd/publisher-example/main.go`
3. Import the package: `import "base-go-app/internal/publisher"`
4. Review the comparison doc: `PUBLISHER_COMPARISON.md`

## Notes

- The Go implementation does NOT include task locking (Laravel's `$exclusive` parameter). If you need task locking, implement it separately using your preferred distributed lock mechanism.
- Task IDs are random UUIDs (v4) vs Laravel's ordered UUIDs. Both are valid and compatible with workers.
- Connection management differs: Go reuses connections, Laravel creates per-task. Both approaches are valid.
