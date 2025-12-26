package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQUser     string
	RabbitMQPassword string
	RabbitMQHost     string
	RabbitMQPort     string
	RabbitMQVHost    string

	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBDatabase string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist, we might be using real env vars
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		RabbitMQUser:     os.Getenv("RABBITMQ_USER"),
		RabbitMQPassword: os.Getenv("RABBITMQ_PASSWORD"),
		RabbitMQHost:     os.Getenv("RABBITMQ_HOST"),
		RabbitMQPort:     os.Getenv("RABBITMQ_PORT"),
		RabbitMQVHost:    os.Getenv("RABBITMQ_VHOST"),

		DBUser:     os.Getenv("DB_USERNAME"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBDatabase: os.Getenv("DB_DATABASE"),
	}

	return cfg, nil
}

func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.RabbitMQUser,
		c.RabbitMQPassword,
		c.RabbitMQHost,
		c.RabbitMQPort,
		c.RabbitMQVHost,
	)
}

func (c *Config) GetDSN() string {
	port := c.DBPort
	if port == "" {
		// Default to the common Postgres port if none was provided to avoid
		// accidental token merging (e.g. "port= sslmode=disable" being
		// parsed incorrectly as the port value).
		port = "5432"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		c.DBHost,
		c.DBUser,
		c.DBPassword,
		c.DBDatabase,
		port,
	)
}
