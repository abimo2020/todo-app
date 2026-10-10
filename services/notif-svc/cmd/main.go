package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/abimo2020/notif-svc/config"
	"github.com/abimo2020/notif-svc/internal/consumer/todocreated"
)

const (
	PORT = 8081
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect rabbitmq: %v", err)
	}
	defer conn.Close()

	consumer, err := todocreated.New(conn)
	if err != nil {
		log.Fatalf("Failed to setup todocreated consumer: %v", err)
	}

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		if err := consumer.Run(context.Background()); err != nil {
			log.Printf("Consumer stopped: %v", err)
			stop()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}

	go func() {
		log.Printf("Server berjalan di :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Println("Sinyal berhenti diterima, memulai shutdown..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("failed to shutdown http server: %v", err)
	}

	select {
	case <-consumerDone:
		log.Println("Consumer done")
	case <-shutdownCtx.Done():
		log.Println("Timeout waiting consumer")
	}
	log.Println("shutdown finish.")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "notif-svc",
	})
}
