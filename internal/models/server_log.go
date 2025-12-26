package models

import (
	"time"

	"github.com/google/uuid"
)

type ServerLog struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key"`
	Message   string                 `gorm:"not null"`
	Channel   string                 `gorm:"not null"`
	Level     int                    `gorm:"not null"`
	LevelName string                 `gorm:"not null"`
	Datetime  string                 `gorm:"not null"`
	Context   map[string]interface{} `gorm:"serializer:json;not null"`
	Extra     map[string]interface{} `gorm:"serializer:json;not null"`
	CreatedAt time.Time              `gorm:"not null"`
	UpdatedAt time.Time              `gorm:"not null"`
}

func (ServerLog) TableName() string {
	return "log"
}
