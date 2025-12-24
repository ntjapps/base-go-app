package tests

import (
	"base-go-app/internal/config"
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"base-go-app/internal/queue"
	"base-go-app/internal/tasks"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestIntegration_LoggerTask(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	// Setup Config
	os.Setenv("RABBITMQ_USER", "guest")
	os.Setenv("RABBITMQ_PASSWORD", "guest")
	os.Setenv("RABBITMQ_HOST", "localhost")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_VHOST", "/")
	os.Setenv("DB_USERNAME", "user")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_DATABASE", "testdb")
	os.Setenv("WORKER_CONCURRENCY", "1")

	cfg, err := config.Load()
	require.NoError(t, err)

	// Connect DB
	err = database.Connect(cfg)
	require.NoError(t, err)
	
	// Migrate
	err = database.DB.AutoMigrate(&models.ServerLog{})
	require.NoError(t, err)

	// Clean DB
	database.DB.Exec("DELETE FROM log")

	// Start Consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := queue.StartConsumer(ctx, cfg)

	// Wait for consumer to connect
	require.Eventually(t, func() bool {
		return queue.RabbitConnected()
	}, 10*time.Second, 100*time.Millisecond)

	// Publish Task
	conn, err := amqp.Dial(cfg.GetRabbitMQURL())
	require.NoError(t, err)
	defer conn.Close()

	ch, err := conn.Channel()
	require.NoError(t, err)
	defer ch.Close()

	// Construct Payload
	payload := tasks.LoggerTaskPayload{
		Message:   "Integration Test Log",
		Channel:   "test",
		Level:     "200",
		LevelName: "INFO",
		Datetime:  time.Now().Format("2006-01-02 15:04:05"),
		Context:   map[string]interface{}{"foo": "bar"},
		Extra:     map[string]interface{}{"baz": "qux"},
	}
	payloadBytes, _ := json.Marshal(payload)

	taskPayload := tasks.TaskPayload{
		Task:    "logger",
		Payload: json.RawMessage(payloadBytes),
	}
	body, _ := json.Marshal(taskPayload)

	err = ch.Publish(
		"celery", // exchange
		"logger", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	require.NoError(t, err)

	// Wait for result in DB
	require.Eventually(t, func() bool {
		var count int64
		database.DB.Model(&models.ServerLog{}).Where("message = ?", "Integration Test Log").Count(&count)
		return count > 0
	}, 10*time.Second, 500*time.Millisecond)

	// Cleanup
	cancel()
	<-done
}
