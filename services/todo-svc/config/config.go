package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	NotifSvcURL string
	DBDsn       string
	RabbitMQURL string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Port:        getEnv("PORT", "8080"),
		NotifSvcURL: getEnv("NOTIF_SVC_URL", "http://localhost:8081"),
		DBDsn:       getEnv("DB_DSN", "postgres://todo_user:todo_pass@localhost:5432/todo_db?sslmode=disable"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://todo_user:todo_pass@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
