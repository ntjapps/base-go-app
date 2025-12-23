package tasks

import (
	"base-go-app/internal/database"
	"base-go-app/internal/models"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type LoggerTaskPayload struct {
	Message   string                 `json:"message"`
	Channel   string                 `json:"channel"`
	Level     string                 `json:"level"`
	LevelName string                 `json:"level_name"`
	Datetime  string                 `json:"datetime"`
	Context   map[string]interface{} `json:"context"`
	Extra     map[string]interface{} `json:"extra"`
}

func HandleLoggerTask(body []byte) error {
	// Try to parse as Celery format first: [args, kwargs, embed]
	var celeryMessage []interface{}
	if err := json.Unmarshal(body, &celeryMessage); err == nil && len(celeryMessage) >= 1 {
		// args is the first element
		if args, ok := celeryMessage[0].([]interface{}); ok && len(args) > 0 {
			// The first arg is our payload
			payloadBytes, _ := json.Marshal(args[0])
			return processPayload(payloadBytes)
		}
	}

	// If not Celery format, try to parse directly
	return processPayload(body)
}

func processPayload(data []byte) error {
	var payload LoggerTaskPayload
	
	// Try to unmarshal directly
	if err := json.Unmarshal(data, &payload); err != nil {
		// Maybe it's a JSON string?
		var strPayload string
		if err2 := json.Unmarshal(data, &strPayload); err2 == nil {
			if err3 := json.Unmarshal([]byte(strPayload), &payload); err3 != nil {
				log.Printf("Error parsing payload string: %v", err3)
				return err3
			}
		} else {
			log.Printf("Error parsing payload: %v", err)
			return err
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to V4 if V7 fails (unlikely)
		id = uuid.New()
	}

	levelInt, _ := strconv.Atoi(payload.Level)

	// Parse datetime
	// Python format: "%Y-%m-%d %H:%M:%S.%f" (but helper does .%03d)
	// Go layout: "2006-01-02 15:04:05.000"
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000", payload.Datetime)
	if err != nil {
		log.Printf("Error parsing time: %v, using Now()", err)
		parsedTime = time.Now()
	}

	logEntry := models.ServerLog{
		ID:        id,
		Message:   payload.Message,
		Channel:   payload.Channel,
		Level:     levelInt,
		LevelName: payload.LevelName,
		Datetime:  parsedTime,
		Context:   payload.Context,
		Extra:     payload.Extra,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result := database.DB.Create(&logEntry)
	if result.Error != nil {
		log.Printf("Error inserting log: %v", result.Error)
	} else {
		log.Printf("Log inserted: %s", id)
	}
	return result.Error
}
