package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	notifHandler "github.com/abimo2020/notif-svc/internal/handler/notification"
)

const (
	PORT = 8081
)

func main() {
	notifHandler := notifHandler.New()

	mux := http.NewServeMux()

	notifHandler.RegisterRoutes(mux)

	mux.HandleFunc("/health", healthHandler)

	log.Printf("Server berjalan di :%d", PORT)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), mux); err != nil {
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
