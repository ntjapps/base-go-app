package tasks

import (
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	database.SetDBForTests(db)
	err = database.DB.AutoMigrate(&models.ServerLog{})
	require.NoError(t, err)
}

func TestLoggerTaskHandler_Handle(t *testing.T) {
	setupTestDB(t)

	handler := &LoggerTaskHandler{}
	
	// Use a map for payload to simulate flexible types (like string level)
	payload := map[string]interface{}{
		"message":    "Test Message",
		"channel":    "test",
		"level":      "200",
		"level_name": "INFO",
		"datetime":   "2023-01-01 12:00:00",
		"context":    map[string]interface{}{"key": "value"},
		"extra":      map[string]interface{}{"foo": "bar"},
	}
	
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), json.RawMessage(payloadBytes))
	assert.NoError(t, err)

	var logEntry models.ServerLog
	err = database.DB.First(&logEntry).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Message", logEntry.Message)
	assert.Equal(t, 200, logEntry.Level)
	assert.Equal(t, "INFO", logEntry.LevelName)
}

func TestLoggerTaskHandler_Handle_EmptyArrayContext(t *testing.T) {
	setupTestDB(t)
	handler := &LoggerTaskHandler{}
	
	// Simulate PHP sending empty array [] for context/extra instead of object {}
	payload := map[string]interface{}{
		"message":    "Empty Context",
		"channel":    "test",
		"level":      200,
		"level_name": "INFO",
		"datetime":   "2023-01-01 12:00:00",
		"context":    []interface{}{}, // Empty array
		"extra":      []interface{}{}, // Empty array
	}
	
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	err = handler.Handle(context.Background(), json.RawMessage(payloadBytes))
	assert.NoError(t, err)

	var logEntry models.ServerLog
	err = database.DB.Where("message = ?", "Empty Context").First(&logEntry).Error
	assert.NoError(t, err)
	assert.Empty(t, logEntry.Context)
	assert.Empty(t, logEntry.Extra)
}

func TestLoggerTaskHandler_Handle_InvalidJSON(t *testing.T) {
	setupTestDB(t)
	handler := &LoggerTaskHandler{}
	
	err := handler.Handle(context.Background(), json.RawMessage(`{invalid`))
	assert.Error(t, err)
}

func TestLoggerTaskHandler_Handle_DBNotConnected(t *testing.T) {
	// Ensure DB is not connected; handler should not panic and should return nil
	database.ClearDBForTests()
	handler := &LoggerTaskHandler{}
	payload := map[string]interface{}{
		"message":    "No DB",
		"channel":    "test",
		"level":      100,
		"level_name": "DEBUG",
		"datetime":   "2023-01-01 12:00:00",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	
	err = handler.Handle(context.Background(), json.RawMessage(payloadBytes))
	assert.NoError(t, err)
}
