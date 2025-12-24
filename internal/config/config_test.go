package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Set environment variables
	os.Setenv("RABBITMQ_USER", "user")
	os.Setenv("RABBITMQ_PASSWORD", "pass")
	os.Setenv("RABBITMQ_HOST", "localhost")
	os.Setenv("RABBITMQ_PORT", "5672")
	os.Setenv("RABBITMQ_VHOST", "vhost")
	os.Setenv("DB_USERNAME", "dbuser")
	os.Setenv("DB_PASSWORD", "dbpass")
	os.Setenv("DB_HOST", "dbhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_DATABASE", "dbname")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "user", cfg.RabbitMQUser)
	assert.Equal(t, "pass", cfg.RabbitMQPassword)
	assert.Equal(t, "localhost", cfg.RabbitMQHost)
	assert.Equal(t, "5672", cfg.RabbitMQPort)
	assert.Equal(t, "vhost", cfg.RabbitMQVHost)

	assert.Equal(t, "dbuser", cfg.DBUser)
	assert.Equal(t, "dbpass", cfg.DBPassword)
	assert.Equal(t, "dbhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "dbname", cfg.DBDatabase)
}

func TestGetRabbitMQURL(t *testing.T) {
	cfg := &Config{
		RabbitMQUser:     "user",
		RabbitMQPassword: "pass",
		RabbitMQHost:     "localhost",
		RabbitMQPort:     "5672",
		RabbitMQVHost:    "vhost",
	}

	expected := "amqp://user:pass@localhost:5672/vhost"
	assert.Equal(t, expected, cfg.GetRabbitMQURL())
}

func TestGetDSN(t *testing.T) {
	cfg := &Config{
		DBUser:     "dbuser",
		DBPassword: "dbpass",
		DBHost:     "dbhost",
		DBPort:     "5432",
		DBDatabase: "dbname",
	}

	expected := "host=dbhost user=dbuser password=dbpass dbname=dbname port=5432 sslmode=disable TimeZone=UTC"
	assert.Equal(t, expected, cfg.GetDSN())
}
