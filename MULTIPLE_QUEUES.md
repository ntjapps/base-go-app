# Multiple Queue Usage Guide

## Overview

The publisher functions fully support sending tasks to **any queue** you specify, enabling parallel processing across multiple workers consuming different queues.

## How It Works

Both `SendCeleryTask()` and `SendGoTask()` accept a `queue` parameter:

```go
// Send to any queue you want
pub.SendGoTask("logger", payload, "go.logger", nil)
pub.SendGoTask("update_data", payload, "go.data_processor", nil)
pub.SendGoTask("send_email", payload, "go.email", nil)
```

## Architecture

```
                                RabbitMQ
                                   │
                    ┌──────────────┼──────────────┐
                    │              │              │
              ┌─────▼─────┐  ┌────▼─────┐  ┌────▼─────┐
              │ go.logger │  │ go.email │  │ go.orders│
              └─────┬─────┘  └────┬─────┘  └────┬─────┘
                    │              │              │
              ┌─────▼─────┐  ┌────▼─────┐  ┌────▼─────┐
              │  Worker 1 │  │ Worker 2 │  │ Worker 3 │
              │  (logger) │  │  (email) │  │ (orders) │
              └───────────┘  └──────────┘  └──────────┘
                    │              │              │
                    └──────────────┴──────────────┘
                           Parallel Processing
```

## Queue Naming Convention

Recommended naming pattern: `{backend}.{purpose}`

- `go.logger` - Go worker for logging
- `go.data_processor` - Go worker for data updates
- `go.email` - Go worker for email sending
- `go.orders` - Go worker for order processing
- `python.analytics` - Python Celery worker for analytics
- `python.ml` - Python Celery worker for ML tasks

## Examples

### Example 1: Separate Queues by Task Type

```go
pub, _ := publisher.NewPublisher(cfg)
defer pub.Close()

// Logger tasks → go.logger queue
pub.SendGoTask("logger", map[string]interface{}{
    "message": "User logged in",
    "level": "200",
}, "go.logger", nil)

// Data updates → go.data_processor queue
pub.SendGoTask("update_user", map[string]interface{}{
    "user_id": 123,
    "status": "active",
}, "go.data_processor", nil)

// Emails → go.email queue
pub.SendGoTask("send_email", map[string]interface{}{
    "to": "user@example.com",
    "subject": "Welcome",
}, "go.email", nil)
```

### Example 2: Priority-Based Queues

```go
// High priority (dedicated fast workers)
pub.SendGoTask("process_payment", payload, "go.high_priority", nil)

// Normal priority (standard workers)
pub.SendGoTask("update_profile", payload, "go.normal_priority", nil)

// Low priority (fewer workers)
pub.SendGoTask("cleanup_temp_files", payload, "go.low_priority", nil)
```

### Example 3: Workload Distribution

```go
// Heavy CPU tasks → dedicated queue with fewer workers
pub.SendGoTask("resize_image", payload, "go.cpu_intensive", &publisher.TaskOptions{
    TimeoutSeconds: ptr(600), // 10 minutes
})

// Light I/O tasks → queue with many workers
pub.SendGoTask("fetch_api_data", payload, "go.io_bound", nil)
```

### Example 4: Multi-Tenant Queues

```go
// Tenant-specific queues for resource isolation
pub.SendGoTask("process_data", payload, "go.tenant_abc", nil)
pub.SendGoTask("process_data", payload, "go.tenant_xyz", nil)
```

### Example 5: Batch Processing Across Multiple Queues

```go
var wg sync.WaitGroup

// Send tasks to different queues concurrently
queues := map[string]map[string]interface{}{
    "go.logger": {"message": "Log entry"},
    "go.email": {"to": "user@example.com"},
    "go.orders": {"order_id": "ORD-123"},
}

for queue, payload := range queues {
    wg.Add(1)
    go func(q string, p map[string]interface{}) {
        defer wg.Done()
        pub.SendGoTask("process", p, q, nil)
    }(queue, payload)
}

wg.Wait()
```

## Worker Configuration

Each worker should be configured to listen to specific queue(s):

### Worker 1 - Logger (Fast, High Throughput)
```bash
# Listen to go.logger queue
RABBITMQ_QUEUE=go.logger go run cmd/worker/main.go
```

### Worker 2 - Data Processor
```bash
# Listen to go.data_processor queue
RABBITMQ_QUEUE=go.data_processor go run cmd/worker/main.go
```

### Worker 3 - Email Sender
```bash
# Listen to go.email queue
RABBITMQ_QUEUE=go.email go run cmd/worker/main.go
```

## Benefits of Multiple Queues

### 1. Parallel Processing
Different workers process different task types simultaneously
```
go.logger → Worker 1 (logging)     } All running
go.email → Worker 2 (emailing)     } in parallel
go.orders → Worker 3 (orders)      }
```

### 2. Resource Isolation
Heavy tasks don't block light tasks
```
go.cpu_intensive → 2 workers (heavy CPU)
go.io_bound → 10 workers (light I/O)
```

### 3. Priority Management
Critical tasks get processed first
```
go.critical → Many fast workers
go.background → Fewer workers
```

### 4. Fault Isolation
Issues in one queue don't affect others
```
go.image_processor crashes
→ go.logger continues working fine
```

### 5. Scalability
Scale workers independently per queue
```
Black Friday traffic:
- Scale up go.orders workers to 50
- Keep go.logger at 10 workers
```

## Complete Example

```go
package main

import (
    "log"
    "base-go-app/internal/config"
    "base-go-app/internal/publisher"
)

func main() {
    cfg, _ := config.Load()
    pub, _ := publisher.NewPublisher(cfg)
    defer pub.Close()

    // User registration flow - distribute across queues
    userID := 12345

    // 1. Log the registration (fast, go.logger)
    pub.SendGoTask("logger", map[string]interface{}{
        "message": "User registered",
        "user_id": userID,
        "level": "200",
    }, "go.logger", nil)

    // 2. Send welcome email (go.email)
    pub.SendGoTask("send_email", map[string]interface{}{
        "to": "user@example.com",
        "template": "welcome",
        "user_id": userID,
    }, "go.email", nil)

    // 3. Update analytics (go.analytics)
    pub.SendGoTask("track_event", map[string]interface{}{
        "event": "user_registered",
        "user_id": userID,
    }, "go.analytics", nil)

    // 4. Generate profile image (slow, go.image_processor)
    timeout := 300
    pub.SendGoTask("generate_avatar", map[string]interface{}{
        "user_id": userID,
    }, "go.image_processor", &publisher.TaskOptions{
        TimeoutSeconds: &timeout,
    })

    log.Println("All registration tasks dispatched to parallel queues!")
}
```

## Queue Declaration

The publisher automatically declares queues as **durable** when sending tasks. If you need custom queue configuration, declare them in advance:

```go
// In your worker initialization or setup script
ch.QueueDeclare(
    "go.logger",  // name
    true,         // durable
    false,        // delete when unused
    false,        // exclusive
    false,        // no-wait
    nil,          // arguments
)
```

## Monitoring

Monitor queue depths to ensure proper resource allocation:

```bash
# Check queue depths
rabbitmqctl list_queues name messages consumers

# Output:
# go.logger           5      10    (5 messages, 10 workers)
# go.email            234    2     (234 messages, 2 workers - may need scaling)
# go.orders           0      5     (0 messages, 5 workers - idle)
```

## Best Practices

1. **Descriptive Names** - Use clear, purpose-driven queue names
2. **Consistent Naming** - Follow a convention (e.g., `{backend}.{purpose}`)
3. **Right-Sized Workers** - Match worker count to queue workload
4. **Monitor Queue Depth** - Scale workers when queues grow
5. **Isolate Heavy Tasks** - Use dedicated queues for CPU/memory-intensive work
6. **Test Failure Scenarios** - Ensure one queue's failure doesn't cascade
7. **Document Queues** - Keep a registry of queues and their purposes

## Summary

✅ **Both functions support any queue name**
- `SendCeleryTask(task, args, queue)` → queue parameter
- `SendGoTask(task, payload, queue, options)` → queue parameter

✅ **No modifications needed** - Feature already implemented!

✅ **Full flexibility** - Send any task to any queue for parallel processing

✅ **Production ready** - Queue isolation, parallel processing, independent scaling
