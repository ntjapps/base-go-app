package publisher

import (
	"encoding/json"
	"fmt"
	"time"

	"base-go-app/internal/config"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements the Publisher interface
type RabbitMQPublisher struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	config *config.Config
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(cfg *config.Config) (*RabbitMQPublisher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	conn, err := amqp.Dial(cfg.GetRabbitMQURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &RabbitMQPublisher{
		conn:   conn,
		ch:     ch,
		config: cfg,
	}, nil
}

// SendCeleryTask sends a task in Celery protocol v2 format (for Python workers)
// This matches the Laravel CeleryFunction trait behavior
func (p *RabbitMQPublisher) SendCeleryTask(task string, args []interface{}, queue string) (string, error) {
	if task == "" {
		return "", fmt.Errorf("task name is required")
	}
	if args == nil {
		args = []interface{}{}
	}
	if queue == "" {
		queue = "celery"
	}

	// Generate task ID
	taskID := uuid.New().String()

	// Declare queue (durable)
	_, err := p.ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue: %w", err)
	}

	// Generate Celery Payload Message Protocol v2
	// Format: [[args...], {kwargs}, {metadata}]
	body := []interface{}{
		args,
		map[string]interface{}{}, // empty kwargs
		map[string]interface{}{   // metadata
			"callbacks": nil,
			"errbacks":  nil,
			"chain":     nil,
			"chord":     nil,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message body: %w", err)
	}

	// Prepare message with Celery headers
	msg := amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		DeliveryMode:    amqp.Persistent,
		Body:            bodyBytes,
		Headers: amqp.Table{
			"lang":    "py",
			"task":    task,
			"id":      taskID,
			"root_id": taskID,
		},
		CorrelationId: taskID,
	}

	// Publish to exchange "celery" with routing key = queue
	err = p.ch.Publish(
		"celery", // exchange
		queue,    // routing key
		false,    // mandatory
		false,    // immediate
		msg,
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	return taskID, nil
}

// SendGoTask sends a task in Go worker format
// This matches the Laravel GoWorkerFunction trait behavior
func (p *RabbitMQPublisher) SendGoTask(task string, payload map[string]interface{}, queue string, options *TaskOptions) (string, error) {
	if task == "" {
		return "", fmt.Errorf("task name is required")
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	if queue == "" {
		queue = "celery"
	}

	// Generate task ID
	taskID := uuid.New().String()

	// Declare queue (durable)
	_, err := p.ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return "", fmt.Errorf("failed to declare queue: %w", err)
	}

	// Build task payload
	taskPayload := map[string]interface{}{
		"version":      "1.0",
		"id":           taskID,
		"task":         task,
		"payload":      payload,
		"created_at":   time.Now().Format(time.RFC3339),
		"attempt":      0,
		"max_attempts": 5,
	}

	// Apply options if provided
	if options != nil {
		if options.TimeoutSeconds != nil {
			taskPayload["timeout_seconds"] = *options.TimeoutSeconds
		}
		if options.Notify != nil {
			taskPayload["notify"] = options.Notify
		}
		if options.MaxAttempts != nil {
			taskPayload["max_attempts"] = *options.MaxAttempts
		}
	}

	bodyBytes, err := json.Marshal(taskPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// Prepare message
	msg := amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf-8",
		DeliveryMode:    amqp.Persistent,
		Body:            bodyBytes,
	}

	// Publish to default exchange (direct to queue)
	err = p.ch.Publish(
		"",    // exchange (empty = default)
		queue, // routing key
		false, // mandatory
		false, // immediate
		msg,
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	return taskID, nil
}

// Close closes the RabbitMQ connection and channel
func (p *RabbitMQPublisher) Close() error {
	var chErr, connErr error

	if p.ch != nil {
		chErr = p.ch.Close()
	}
	if p.conn != nil {
		connErr = p.conn.Close()
	}

	if chErr != nil {
		return fmt.Errorf("failed to close channel: %w", chErr)
	}
	if connErr != nil {
		return fmt.Errorf("failed to close connection: %w", connErr)
	}

	return nil
}
