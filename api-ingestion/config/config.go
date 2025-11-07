package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	RabbitMQURL string
}

func Load() *Config {
	// Try to load .env file (for local development)
	// In Docker, environment variables are injected directly
	_ = godotenv.Load("../.env")

	return &Config{
		Port:        getEnv("API_INGESTION_PORT", "8080"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
