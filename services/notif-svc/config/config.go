package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		Port: getEnv("PORT", "8081"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
