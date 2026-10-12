package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	notifClient, err := notif.NewRabbitMQ(rabbitMQConn)
	if err != nil {
		log.Fatalf("failed to setup notif client: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	todoRepo := todoRepo.New(db)
	todoSvc := todoSvc.New(todoRepo, notifClient)
	todoHandler := todo.New(todoSvc)

	todoHandler.RegisterRoutes(mux)

	srv := http.Server{Addr: ":" + cfg.Port, Handler: mux}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server berjalan di :%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Receive stop signal, starting shutdown..")
	case err := <-serverErr:
		log.Printf("http server error: %v", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}
	log.Println("Success to shutdown")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "todo-svc",
	})
}
