package tasks

import (
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// LoggerTaskHandler implements TaskHandler for logging tasks.
type LoggerTaskHandler struct{}

// LoggerTaskPayload defines the expected structure of the logger task arguments.
type LoggerTaskPayload struct {
	Message   string                 `json:"message"`
	Channel   string                 `json:"channel"`
	Level     interface{}            `json:"level"` // Can be int or string
	LevelName string                 `json:"level_name"`
	Datetime  string                 `json:"datetime"`
	Context   interface{}            `json:"context"` // Can be map or array (empty array in PHP = [])
	Extra     interface{}            `json:"extra"`   // Can be map or array (empty array in PHP = [])
}

// Handle processes the logger task.
func (h *LoggerTaskHandler) Handle(ctx context.Context, args json.RawMessage) error {
	var payload LoggerTaskPayload
	if err := json.Unmarshal(args, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal logger payload: %w", err)
	}

	return processLoggerPayload(payload)
}

func processLoggerPayload(payload LoggerTaskPayload) error {
	// Convert Level to int
	var levelInt int
	var err error

	switch v := payload.Level.(type) {
	case float64:
		levelInt = int(v)
	case int:
		levelInt = v
	case string:
		levelInt, err = strconv.Atoi(v)
		if err != nil {
			levelInt = 0
			log.Printf("Warning: invalid level string %s, defaulting to 0", v)
		}
	default:
		levelInt = 0
		log.Printf("Warning: invalid level type %T, defaulting to 0", v)
	}

	// Normalize Context and Extra to map[string]interface{}
	contextMap := normalizeToMap(payload.Context)
	extraMap := normalizeToMap(payload.Extra)

	// Parse Datetime
	// We'll try standard formats.
	logDate, err := time.Parse("2006-01-02 15:04:05.000000", payload.Datetime)
	if err != nil {
		// Try without microseconds
		logDate, err = time.Parse("2006-01-02 15:04:05", payload.Datetime)
		if err != nil {
			// Try with 3 digits ms
			logDate, err = time.Parse("2006-01-02 15:04:05.000", payload.Datetime)
			if err != nil {
				logDate = time.Now()
				log.Printf("Warning: invalid datetime %s, defaulting to now", payload.Datetime)
			}
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		id = uuid.New()
	}

	serverLog := models.ServerLog{
		ID:        id,
		Message:   payload.Message,
		Channel:   payload.Channel,
		Level:     levelInt,
		LevelName: payload.LevelName,
		Datetime:  logDate.Format("2006-01-02 15:04:05.000000"),
		Context:   contextMap,
		Extra:     extraMap,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// If DB is not connected, skip persisting logs to avoid panics and
	// allow the worker to continue processing other tasks.
	if !database.Connected() || database.DB == nil {
		log.Printf("Database not connected; skipping saving log: %s", serverLog.ID)
		return nil
	}

	if err := database.DB.Create(&serverLog).Error; err != nil {
		log.Printf("Failed to save log to DB: %v", err)
		return err
	}

	log.Printf("Successfully saved log: %s", serverLog.ID)
	return nil
}

// normalizeToMap converts interface{} to map[string]interface{}.
// Handles empty arrays (which PHP sends for empty maps) by returning an empty map.
func normalizeToMap(v interface{}) map[string]interface{} {
	if v == nil {
		return make(map[string]interface{})
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	// If it's an array (likely empty from PHP), return empty map
	if _, ok := v.([]interface{}); ok {
		return make(map[string]interface{})
	}
	// Fallback: try to marshal/unmarshal if it's some other structure, or return empty
	return make(map[string]interface{})
}

// Register the handler
func init() {
	RegisterTask("logger", &LoggerTaskHandler{})
}
