package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/abimo2020/todo-svc/config"
	"github.com/abimo2020/todo-svc/internal/gateway/notif"
	"github.com/abimo2020/todo-svc/internal/handler/todo"
	todoRepo "github.com/abimo2020/todo-svc/internal/repository/todo"
	todoSvc "github.com/abimo2020/todo-svc/internal/service/todo"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	rabbitMQConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect rabbitMQ: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	// Notif client using http protocol
	// notifClient := notif.New(cfg.NotifSvcURL)

	// Notif client using rabbitMQ
	notifClient, err := notif.NewRabbitMQ(rabbitMQConn)
	if err != nil {
		log.Fatalf("failed to setup notif client: %v", err)
	}

	todoRepo := todoRepo.New(db)
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
