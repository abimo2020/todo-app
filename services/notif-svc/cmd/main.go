package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/abimo2020/notif-svc/config"
	"github.com/abimo2020/notif-svc/internal/consumer/todocreated"
)

const (
	PORT = 8081
)

func main() {
	cfg := config.Load()

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect rabbitmq: %v", err)
	}
	defer conn.Close()

	consumer, err := todocreated.New(conn)
	if err != nil {
		log.Fatalf("Failed to setup todocreated consumer: %v", err)
	}

	go func() {
		if err := consumer.Run(context.Background()); err != nil {
			log.Fatalf("Consumer stopped: %v", err)
		}
	}()

	mux := http.NewServeMux()

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
