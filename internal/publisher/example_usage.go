package publisher

// Example usage demonstrating how to use the publisher package
// This file is for documentation purposes only

/*
import (
    "log"
    "base-go-app/internal/config"
    "base-go-app/internal/publisher"
)

func ExampleUsage() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Create publisher instance
    pub, err := publisher.NewPublisher(cfg)
    if err != nil {
        log.Fatalf("Failed to create publisher: %v", err)
    }
    defer pub.Close()

    // Example 1: Send a Celery task (for Python Celery workers)
    // This is compatible with Laravel CeleryFunction::sendTask()
    taskID, err := pub.SendCeleryTask(
        "celery_test_task",           // task name
        []interface{}{"arg1", "arg2"}, // arguments
        "celery",                      // queue name
    )
    if err != nil {
        log.Fatalf("Failed to send Celery task: %v", err)
    }
    log.Printf("Celery task submitted with ID: %s", taskID)

    // Example 2: Send a Celery task with multiple arguments
    taskID, err = pub.SendCeleryTask(
        "celery_test_body_task",
        []interface{}{"value1", "value2", "value3"},
        "celery",
    )
    if err != nil {
        log.Fatalf("Failed to send Celery task: %v", err)
    }
    log.Printf("Celery body task submitted with ID: %s", taskID)

    // Example 3: Send a Go worker task (for Go workers)
    // This is compatible with Laravel GoWorkerFunction::sendGoTask()
    payload := map[string]interface{}{
        "message":    "Test log message",
        "channel":    "app",
        "level":      "200",
        "level_name": "INFO",
        "datetime":   "2025-12-30 10:00:00",
        "context":    map[string]interface{}{"user_id": 123},
        "extra":      map[string]interface{}{"request_id": "abc-123"},
    }

    taskID, err = pub.SendGoTask(
        "logger",  // task name
        payload,   // payload data
        "logger",  // queue name
        nil,       // no options
    )
    if err != nil {
        log.Fatalf("Failed to send Go task: %v", err)
    }
    log.Printf("Go task submitted with ID: %s", taskID)

    // Example 4: Send a Go worker task with options
    timeout := 300
    maxAttempts := 3
    options := &publisher.TaskOptions{
        TimeoutSeconds: &timeout,
        Notify: map[string]string{
            "webhook": "http://example.com/callback",
            "email":   "admin@example.com",
        },
        MaxAttempts: &maxAttempts,
    }

    payload = map[string]interface{}{
        "message":    "Important notification",
        "channel":    "notifications",
        "level":      "300",
        "level_name": "WARNING",
    }

    taskID, err = pub.SendGoTask(
        "notification",
        payload,
        "notifications",
        options,
    )
    if err != nil {
        log.Fatalf("Failed to send notification task: %v", err)
    }
    log.Printf("Notification task submitted with ID: %s", taskID)

    // Example 5: Reusing the publisher for multiple tasks
    for i := 0; i < 10; i++ {
        taskID, err := pub.SendGoTask(
            "logger",
            map[string]interface{}{
                "message": fmt.Sprintf("Batch log %d", i),
                "level":   "200",
            },
            "logger",
            nil,
        )
        if err != nil {
            log.Printf("Failed to send batch task %d: %v", i, err)
            continue
        }
        log.Printf("Batch task %d submitted with ID: %s", i, taskID)
    }
}
*/
