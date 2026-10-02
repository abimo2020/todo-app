package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/abimo2020/todo-svc/internal/gateway/notif"
	"github.com/abimo2020/todo-svc/internal/handler/todo"
	todoRepo "github.com/abimo2020/todo-svc/internal/repository/todo"
	todoSvc "github.com/abimo2020/todo-svc/internal/service/todo"
)

const (
	PORT = 8080
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	notifClient := notif.New("http://localhost:8081")

	todoRepo := todoRepo.New()
	todoSvc := todoSvc.New(todoRepo, notifClient)
	todoHandler := todo.New(todoSvc)

	todoHandler.RegisterRoutes(mux)

	log.Printf("Server berjalan di :%d\n", PORT)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), mux); err != nil {
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
