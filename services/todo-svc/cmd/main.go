package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/abimo2020/todo-svc/config"
	"github.com/abimo2020/todo-svc/internal/gateway/notif"
	"github.com/abimo2020/todo-svc/internal/handler/todo"
	todoRepo "github.com/abimo2020/todo-svc/internal/repository/todo"
	todoSvc "github.com/abimo2020/todo-svc/internal/service/todo"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	notifClient := notif.New(cfg.NotifSvcURL)

	todoRepo := todoRepo.New()
	todoSvc := todoSvc.New(todoRepo, notifClient)
	todoHandler := todo.New(todoSvc)

	todoHandler.RegisterRoutes(mux)

	log.Printf("Server berjalan di :%s\n", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "todo-svc",
	})
}
