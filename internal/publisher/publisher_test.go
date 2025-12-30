package publisher

import (
	"encoding/json"
	"testing"
	"time"

	"base-go-app/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublisher(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		pub, err := NewPublisher(nil)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid connection", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.RabbitMQHost = "invalid-host"
		cfg.RabbitMQPort = "5672"
		cfg.RabbitMQUser = "guest"
		cfg.RabbitMQPassword = "guest"
		cfg.RabbitMQVHost = "/"

		pub, err := NewPublisher(cfg)
		assert.Error(t, err)
		assert.Nil(t, pub)
		assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
	})
}

func TestSendCeleryTask(t *testing.T) {
	t.Run("empty task name", func(t *testing.T) {
		// Use mock publisher (validation happens before channel access)
		pub := &RabbitMQPublisher{}
		
		taskID, err := pub.SendCeleryTask("", nil, "celery")
		assert.Error(t, err)
		assert.Empty(t, taskID)
		assert.Contains(t, err.Error(), "task name is required")
	})
}

func TestSendGoTask(t *testing.T) {
	t.Run("empty task name", func(t *testing.T) {
		pub := &RabbitMQPublisher{}
		
		taskID, err := pub.SendGoTask("", nil, "celery", nil)
		assert.Error(t, err)
		assert.Empty(t, taskID)
		assert.Contains(t, err.Error(), "task name is required")
	})
}

func TestClose(t *testing.T) {
	t.Run("close nil connections", func(t *testing.T) {
		pub := &RabbitMQPublisher{}
		err := pub.Close()
		assert.NoError(t, err)
	})
}

// Integration tests (require RabbitMQ to be running)
func TestIntegration_SendCeleryTask(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.Config{
		RabbitMQHost:     "localhost",
		RabbitMQPort:     "5672",
		RabbitMQUser:     "guest",
		RabbitMQPassword: "guest",
		RabbitMQVHost:    "/",
	}

	pub, err := NewPublisher(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
		return
	}
	defer pub.Close()

	t.Run("send celery task successfully", func(t *testing.T) {
		taskID, err := pub.SendCeleryTask(
			"celery_test_task",
			[]interface{}{"arg1", "arg2", 123},
			"test_queue",
		)
		require.NoError(t, err)
		assert.NotEmpty(t, taskID)

		// Verify message was published by consuming it
		msgs, err := pub.ch.Consume(
			"test_queue",
			"",
			true,  // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		require.NoError(t, err)

		select {
		case msg := <-msgs:
			// Verify message structure
			var body []interface{}
			err := json.Unmarshal(msg.Body, &body)
			require.NoError(t, err)
			assert.Len(t, body, 3) // [args, kwargs, metadata]

			// Check headers
			assert.Equal(t, "py", msg.Headers["lang"])
			assert.Equal(t, "celery_test_task", msg.Headers["task"])
			assert.Equal(t, taskID, msg.Headers["id"])
			assert.Equal(t, taskID, msg.CorrelationId)

		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for message")
		}
	})
}

func TestIntegration_SendGoTask(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.Config{
		RabbitMQHost:     "localhost",
		RabbitMQPort:     "5672",
		RabbitMQUser:     "guest",
		RabbitMQPassword: "guest",
		RabbitMQVHost:    "/",
	}

	pub, err := NewPublisher(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
		return
	}
	defer pub.Close()

	t.Run("send go task successfully", func(t *testing.T) {
		timeout := 300
		options := &TaskOptions{
			TimeoutSeconds: &timeout,
			Notify: map[string]string{
				"webhook": "http://example.com/callback",
			},
		}

		payload := map[string]interface{}{
			"message":    "test log message",
			"level":      "info",
			"level_name": "INFO",
		}

		taskID, err := pub.SendGoTask("logger", payload, "test_go_queue", options)
		require.NoError(t, err)
		assert.NotEmpty(t, taskID)

		// Verify message was published by consuming it
		msgs, err := pub.ch.Consume(
			"test_go_queue",
			"",
			true,  // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		require.NoError(t, err)

		select {
		case msg := <-msgs:
			// Verify message structure
			var taskPayload map[string]interface{}
			err := json.Unmarshal(msg.Body, &taskPayload)
			require.NoError(t, err)

			assert.Equal(t, "1.0", taskPayload["version"])
			assert.Equal(t, taskID, taskPayload["id"])
			assert.Equal(t, "logger", taskPayload["task"])
			assert.Equal(t, float64(300), taskPayload["timeout_seconds"])
			assert.NotNil(t, taskPayload["notify"])
			assert.NotNil(t, taskPayload["payload"])
			assert.Equal(t, float64(0), taskPayload["attempt"])
			assert.Equal(t, float64(5), taskPayload["max_attempts"])

		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for message")
		}
	})
}

func TestIntegration_CeleryTaskFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &config.Config{
		RabbitMQHost:     "localhost",
		RabbitMQPort:     "5672",
		RabbitMQUser:     "guest",
		RabbitMQPassword: "guest",
		RabbitMQVHost:    "/",
	}

	pub, err := NewPublisher(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
		return
	}
	defer pub.Close()

	t.Run("celery message format matches Laravel output", func(t *testing.T) {
		// This test ensures compatibility with Python Celery workers
		taskID, err := pub.SendCeleryTask(
			"celery_test_body_task",
			[]interface{}{"value1", "value2", "value3"},
			"celery",
		)
		require.NoError(t, err)

		// Manually consume to verify format
		conn, err := amqp.Dial(cfg.GetRabbitMQURL())
		require.NoError(t, err)
		defer conn.Close()

		ch, err := conn.Channel()
		require.NoError(t, err)
		defer ch.Close()

		msgs, err := ch.Consume("celery", "", false, false, false, false, nil)
		require.NoError(t, err)

		select {
		case msg := <-msgs:
			// Parse body
			var body []interface{}
			err := json.Unmarshal(msg.Body, &body)
			require.NoError(t, err)

			// Verify Celery v2 protocol structure: [args, kwargs, metadata]
			require.Len(t, body, 3)
			
			args, ok := body[0].([]interface{})
			require.True(t, ok)
			assert.Equal(t, []interface{}{"value1", "value2", "value3"}, args)

			kwargs, ok := body[1].(map[string]interface{})
			require.True(t, ok)
			assert.Empty(t, kwargs)

			metadata, ok := body[2].(map[string]interface{})
			require.True(t, ok)
			assert.Nil(t, metadata["callbacks"])
			assert.Nil(t, metadata["errbacks"])
			assert.Nil(t, metadata["chain"])
			assert.Nil(t, metadata["chord"])

			// Ack message
			msg.Ack(false)

		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for message")
		}

		_ = taskID
	})
}
