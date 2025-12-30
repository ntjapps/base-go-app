# RabbitMQ Publisher Functions - PHP vs Go Comparison

This document shows how the Go publisher functions correspond to the Laravel traits for publishing tasks to RabbitMQ.

## Overview

The base-go-app now includes publisher functions that match the functionality of the Laravel traits:
- **Laravel**: `CeleryFunction` and `GoWorkerFunction` traits
- **Go**: `internal/publisher` package

## Function Comparison

### 1. Celery Task (Python Workers)

#### Laravel (CeleryFunction.php)
```php
use App\Traits\CeleryFunction;

class MyController {
    use CeleryFunction;
    
    public function sendTask() {
        $taskId = $this->sendTask(
            'celery_test_task',           // task name
            ['arg1', 'arg2'],             // args
            'celery',                     // queue
            false,                        // exclusive
            null                          // timeout
        );
    }
}
```

#### Go (publisher package)
```go
import "base-go-app/internal/publisher"

func sendTask() {
    pub, _ := publisher.NewPublisher(cfg)
    defer pub.Close()
    
    taskID, err := pub.SendCeleryTask(
        "celery_test_task",           // task name
        []interface{}{"arg1", "arg2"}, // args
        "celery",                     // queue
    )
}
```

### 2. Go Worker Task

#### Laravel (GoWorkerFunction.php)
```php
use App\Traits\GoWorkerFunction;

class MyController {
    use GoWorkerFunction;
    
    public function sendGoTask() {
        $taskId = $this->sendGoTask(
            'logger',                     // task name
            [                             // payload
                'message' => 'test',
                'level' => '200',
            ],
            'logger',                     // queue
            false,                        // exclusive
            300,                          // timeout
            [                             // notify
                'webhook' => 'http://...',
            ]
        );
    }
}
```

#### Go (publisher package)
```go
import "base-go-app/internal/publisher"

func sendGoTask() {
    pub, _ := publisher.NewPublisher(cfg)
    defer pub.Close()
    
    timeout := 300
    options := &publisher.TaskOptions{
        TimeoutSeconds: &timeout,
        Notify: map[string]string{
            "webhook": "http://...",
        },
    }
    
    payload := map[string]interface{}{
        "message": "test",
        "level":   "200",
    }
    
    taskID, err := pub.SendGoTask(
        "logger",   // task name
        payload,    // payload
        "logger",   // queue
        options,    // options
    )
}
```

## Message Format Compatibility

### Celery Protocol v2 (Python Workers)

Both implementations produce identical message formats:

**Message Body:**
```json
[
    ["arg1", "arg2"],              // args array
    {},                            // kwargs (empty)
    {                              // metadata
        "callbacks": null,
        "errbacks": null,
        "chain": null,
        "chord": null
    }
]
```

**Headers:**
```json
{
    "lang": "py",
    "task": "celery_test_task",
    "id": "uuid-here",
    "root_id": "uuid-here"
}
```

**Properties:**
- `content_type`: "application/json"
- `content_encoding`: "utf-8"
- `correlation_id`: task UUID
- `delivery_mode`: persistent

### Go Worker Format

Both implementations produce identical message formats:

```json
{
    "version": "1.0",
    "id": "uuid-here",
    "task": "logger",
    "payload": {
        "message": "test",
        "level": "200"
    },
    "created_at": "2025-12-30T10:00:00Z",
    "attempt": 0,
    "max_attempts": 5,
    "timeout_seconds": 300,
    "notify": {
        "webhook": "http://..."
    }
}
```

## Key Differences

### Connection Management

**Laravel:**
- Opens and closes connection for each task
- Connection details from `config('services.rabbitmq.*)`
- Uses PhpAmqpLib

**Go:**
- Reuses connection for multiple tasks
- Must call `Close()` when done (use `defer`)
- Connection details from environment variables
- Uses rabbitmq/amqp091-go

### Error Handling

**Laravel:**
- Throws `CommonCustomException` on errors
- Task locks managed with Laravel Cache

**Go:**
- Returns error as second return value
- No built-in task locking (implement separately if needed)

### Task ID Generation

**Laravel:**
- Uses `Str::orderedUuid()->toString()`
- Results in ordered UUIDs (time-based)

**Go:**
- Uses `uuid.New().String()`
- Results in random UUIDs (v4)

## Example Use Cases

### 1. Send Log to Database

**Laravel:**
```php
$this->sendGoTask('logger', [
    'message' => 'User logged in',
    'level' => '200',
    'level_name' => 'INFO',
], 'logger');
```

**Go:**
```go
pub.SendGoTask("logger", map[string]interface{}{
    "message":    "User logged in",
    "level":      "200",
    "level_name": "INFO",
}, "logger", nil)
```

### 2. Trigger Python Celery Task

**Laravel:**
```php
$this->sendTask('celery_test_body_task', [
    'value1', 'value2', 'value3'
], 'celery');
```

**Go:**
```go
pub.SendCeleryTask("celery_test_body_task", []interface{}{
    "value1", "value2", "value3",
}, "celery")
```

### 3. Send with Timeout and Notification

**Laravel:**
```php
$this->sendGoTask('notification', [
    'message' => 'Alert!'
], 'notifications', false, 300, [
    'webhook' => 'http://example.com/callback'
]);
```

**Go:**
```go
timeout := 300
pub.SendGoTask("notification", map[string]interface{}{
    "message": "Alert!",
}, "notifications", &publisher.TaskOptions{
    TimeoutSeconds: &timeout,
    Notify: map[string]string{
        "webhook": "http://example.com/callback",
    },
})
```

## Testing

Both implementations include comprehensive tests:

**Laravel:**
- Uses PHPUnit
- Tests are typically integration tests requiring RabbitMQ

**Go:**
- Uses testify
- Unit tests with `-short` flag (no RabbitMQ required)
- Integration tests without `-short` flag (RabbitMQ required)

Run Go tests:
```bash
# Unit tests only
go test ./internal/publisher/... -short

# All tests (including integration)
go test ./internal/publisher/...
```

## Migration Guide

### From Laravel to Go

If you're migrating code from Laravel to Go:

1. **Replace trait usage:**
   ```php
   // Laravel
   use App\Traits\GoWorkerFunction;
   $this->sendGoTask(...)
   ```
   ```go
   // Go
   import "base-go-app/internal/publisher"
   pub, _ := publisher.NewPublisher(cfg)
   defer pub.Close()
   pub.SendGoTask(...)
   ```

2. **Convert payload arrays to maps:**
   ```php
   // Laravel
   ['key' => 'value']
   ```
   ```go
   // Go
   map[string]interface{}{"key": "value"}
   ```

3. **Handle errors explicitly:**
   ```php
   // Laravel
   try {
       $id = $this->sendGoTask(...);
   } catch (CommonCustomException $e) {
       // handle
   }
   ```
   ```go
   // Go
   id, err := pub.SendGoTask(...)
   if err != nil {
       // handle
   }
   ```

## Summary

The Go publisher functions provide complete compatibility with the Laravel traits, ensuring that tasks published from either PHP or Go applications will work with the same worker infrastructure (Python Celery workers or Go workers).

Both implementations:
- ✅ Follow the same message protocols
- ✅ Support the same task types
- ✅ Use the same RabbitMQ configuration
- ✅ Generate compatible task IDs
- ✅ Include comprehensive tests
- ✅ Support task options (timeout, notifications, etc.)
