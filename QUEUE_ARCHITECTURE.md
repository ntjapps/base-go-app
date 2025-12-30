# Queue Architecture Diagram

## Single Publisher, Multiple Queues, Multiple Workers

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Your Go Application                          │
│                                                                     │
│  import "base-go-app/internal/publisher"                           │
│                                                                     │
│  pub, _ := publisher.NewPublisher(cfg)                             │
│  defer pub.Close()                                                 │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ pub.SendGoTask("logger", payload, "go.logger", nil)          │ │
│  │ pub.SendGoTask("update_data", payload, "go.data_processor")  │ │
│  │ pub.SendGoTask("send_email", payload, "go.email")            │ │
│  │ pub.SendCeleryTask("analyze", args, "python.analytics")      │ │
│  └──────────────────────────────────────────────────────────────┘ │
└─────────────────┬───────────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          RabbitMQ Broker                            │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐            │
│  │  go.logger   │  │   go.email   │  │  go.orders   │            │
│  │              │  │              │  │              │            │
│  │ [msg][msg]   │  │ [msg][msg]   │  │ [msg]        │            │
│  │ [msg][msg]   │  │ [msg]        │  │              │            │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘            │
│         │                  │                  │                    │
│  ┌──────┴────────┐  ┌─────┴───────┐  ┌──────┴────────┐           │
│  │go.data_processor│ │go.image_proc│  │python.analytics│          │
│  │              │  │              │  │              │            │
│  │ [msg][msg]   │  │ [msg][msg]   │  │ [msg]        │            │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘            │
└─────────┼──────────────────┼──────────────────┼───────────────────┘
          │                  │                  │
          │                  │                  │
┌─────────▼───────┐  ┌───────▼──────┐  ┌───────▼──────┐
│                 │  │              │  │              │
│  Worker 1-10    │  │  Worker 1-2  │  │  Worker 1-3  │
│                 │  │              │  │              │
│  Queue:         │  │  Queue:      │  │  Queue:      │
│  go.logger      │  │  go.email    │  │  go.orders   │
│                 │  │              │  │              │
│  Task: logger   │  │  Task:       │  │  Task:       │
│  Fast logging   │  │  send_email  │  │  process_order│
│  High volume    │  │  Medium vol. │  │  Low volume  │
└─────────────────┘  └──────────────┘  └──────────────┘

┌─────────────────┐  ┌──────────────┐  ┌──────────────┐
│                 │  │              │  │              │
│  Worker 1-5     │  │  Worker 1-2  │  │  Worker 1-4  │
│                 │  │              │  │ (Python)     │
│  Queue:         │  │  Queue:      │  │              │
│  go.data_proc   │  │  go.image_   │  │  Queue:      │
│                 │  │  processor   │  │  python.     │
│  Task:          │  │              │  │  analytics   │
│  update_data    │  │  Task:       │  │              │
│  DB operations  │  │  resize_image│  │  Task:       │
│                 │  │  CPU-heavy   │  │  generate_   │
│                 │  │  Long tasks  │  │  report      │
└─────────────────┘  └──────────────┘  └──────────────┘
```

## Benefits

### ✅ Parallel Processing
All queues process simultaneously - no blocking between task types

### ✅ Independent Scaling
```
- go.logger:        10 workers (high volume)
- go.email:         2 workers  (medium volume)
- go.image_proc:    2 workers  (CPU-intensive)
- go.data_proc:     5 workers  (database operations)
- python.analytics: 4 workers  (Python-specific)
```

### ✅ Fault Isolation
If image processing workers crash, logging continues unaffected

### ✅ Priority Management
Critical queues get more workers, background queues get fewer

### ✅ Technology Mix
Mix Go workers and Python Celery workers in same infrastructure

## Real-World Example: User Registration

```
User Registers
     │
     ├─ SendGoTask("logger", {...}, "go.logger")
     │  └─► Worker 1 logs immediately (50ms)
     │
     ├─ SendGoTask("send_email", {...}, "go.email")  
     │  └─► Worker 2 sends email (2 seconds)
     │
     ├─ SendGoTask("update_stats", {...}, "go.analytics")
     │  └─► Worker 3 updates DB (100ms)
     │
     └─ SendGoTask("generate_avatar", {...}, "go.image_processor")
        └─► Worker 4 creates image (5 seconds)

All happen in PARALLEL!
Total time: ~5 seconds (not 5s + 2s + 100ms + 50ms = 7.15s)
```

## Queue Naming Convention

```
{backend}.{purpose}[.{priority}]

Examples:
- go.logger              → Go worker, logging
- go.logger.high         → Go worker, high-priority logging
- go.data_processor      → Go worker, data processing
- go.email               → Go worker, email sending
- python.analytics       → Python worker, analytics
- python.ml.gpu          → Python worker, ML with GPU
```

## Monitoring Commands

```bash
# List all queues with message counts
rabbitmqctl list_queues name messages consumers

# Watch specific queue
watch -n 1 'rabbitmqctl list_queues name messages | grep go.logger'

# Check queue bindings
rabbitmqadmin list bindings

# Purge a queue (development only!)
rabbitmqctl purge_queue go.logger
```

## Summary

✅ One publisher can send to unlimited queues
✅ Each queue can have dedicated workers
✅ Perfect for parallel processing
✅ Easy to scale per workload type
✅ Mix Go and Python workers seamlessly
