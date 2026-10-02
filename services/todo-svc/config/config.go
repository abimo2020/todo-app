package config

import "os"

type Config struct {
	Port        string
	NotifSvcURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		NotifSvcURL: getEnv("NOTIF_SVC_URL", "http://localhost:8081"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
