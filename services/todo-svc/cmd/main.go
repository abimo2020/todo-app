package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
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
	"github.com/abimo2020/todo-svc/internal/middleware"
	todoRepo "github.com/abimo2020/todo-svc/internal/repository/todo"
	todoSvc "github.com/abimo2020/todo-svc/internal/service/todo"
)

func main() {
	cfg := config.Load()

	setupLogger(cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect db: %v", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error(fmt.Sprintf("failed to ping db: %v", err))
		os.Exit(1)
	}

	rabbitMQConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to connect rabbitMQ: %v", err))
		os.Exit(1)
	}

	notifClient, err := notif.NewRabbitMQ(rabbitMQConn)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to setup notif client: %v", err))
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	todoRepo := todoRepo.New(db)
	todoSvc := todoSvc.New(todoRepo, notifClient)
	todoHandler := todo.New(todoSvc)

	todoHandler.RegisterRoutes(mux)

	srv := http.Server{Addr: ":" + cfg.Port, Handler: middleware.Logging(mux)}

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

func setupLogger(format string) {
	opts := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &opts)
	}

	slog.SetDefault(slog.New(handler))
}
