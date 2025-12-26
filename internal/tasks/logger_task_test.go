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
	
	payload := LoggerTaskPayload{
		Message:   "Test Message",
		Channel:   "test",
		Level:     "200",
		LevelName: "INFO",
		Datetime:  "2023-01-01 12:00:00",
		Context:   map[string]interface{}{"key": "value"},
		Extra:     map[string]interface{}{"foo": "bar"},
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
	payload := LoggerTaskPayload{
		Message:   "No DB",
		Channel:   "test",
		Level:     "100",
		LevelName: "DEBUG",
		Datetime:  "2023-01-01 12:00:00",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	
	err = handler.Handle(context.Background(), json.RawMessage(payloadBytes))
	assert.NoError(t, err)
}
