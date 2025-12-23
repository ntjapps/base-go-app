package helpers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LogPayload struct {
	ID        string                 `json:"id"`
	Message   string                 `json:"message"`
	Channel   string                 `json:"channel"`
	Level     string                 `json:"level"`
	LevelName string                 `json:"level_name"`
	Datetime  string                 `json:"datetime"`
	Context   map[string]interface{} `json:"context"`
	Extra     map[string]interface{} `json:"extra"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

func LaravelLogPayload(message string, level string, context map[string]interface{}, extra map[string]interface{}) LogPayload {
	levelName := strings.ToUpper(level)
	levelNum := 200

	switch strings.ToLower(level) {
	case "debug":
		levelNum = 100
	case "info":
		levelNum = 200
	case "notice":
		levelNum = 250
	case "warning":
		levelNum = 300
	case "error":
		levelNum = 400
	case "critical":
		levelNum = 500
	case "alert":
		levelNum = 550
	case "emergency":
		levelNum = 600
	}

	now := time.Now()
	datetimeStr := now.Format("2006-01-02 15:04:05.000")

	if context == nil {
		context = make(map[string]interface{})
	}
	if extra == nil {
		extra = make(map[string]interface{})
	}

	id, _ := uuid.NewV7()

	return LogPayload{
		ID:        id.String(),
		Message:   message,
		Channel:   "celery",
		Level:     fmt.Sprintf("%d", levelNum),
		LevelName: levelName,
		Datetime:  datetimeStr,
		Context:   context,
		Extra:     extra,
		CreatedAt: datetimeStr,
		UpdatedAt: datetimeStr,
	}
}
