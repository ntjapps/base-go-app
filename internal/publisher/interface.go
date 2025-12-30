package publisher

// Publisher defines the interface for publishing tasks to RabbitMQ
type Publisher interface {
	// SendCeleryTask sends a task in Celery protocol v2 format (Python workers)
	// task: task name (e.g., "celery_test_task")
	// args: array of arguments for the task
	// queue: RabbitMQ queue name (default: "celery")
	SendCeleryTask(task string, args []interface{}, queue string) (string, error)

	// SendGoTask sends a task in Go worker format
	// task: task name (e.g., "logger")
	// payload: map of task payload data
	// queue: RabbitMQ queue name (default: "celery")
	// options: optional task options (timeout, notify, etc.)
	SendGoTask(task string, payload map[string]interface{}, queue string, options *TaskOptions) (string, error)

	// Close closes the RabbitMQ connection
	Close() error
}

// TaskOptions contains optional parameters for Go tasks
type TaskOptions struct {
	TimeoutSeconds *int               `json:"timeout_seconds,omitempty"`
	Notify         map[string]string  `json:"notify,omitempty"`
	MaxAttempts    *int               `json:"max_attempts,omitempty"`
}
