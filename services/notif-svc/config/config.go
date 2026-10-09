package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	RabbitMQURL string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		Port:        getEnv("PORT", "8081"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://todo_user:todo_pass@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
