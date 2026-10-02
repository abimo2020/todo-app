package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/abimo2020/notif-svc/config"
	notifHandler "github.com/abimo2020/notif-svc/internal/handler/notification"
)

const (
	PORT = 8081
)

func main() {
	cfg := config.Load()

	notifHandler := notifHandler.New()

	mux := http.NewServeMux()

	notifHandler.RegisterRoutes(mux)

	mux.HandleFunc("/health", healthHandler)

	log.Printf("Server berjalan di :%s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "notif-svc",
	})
}
