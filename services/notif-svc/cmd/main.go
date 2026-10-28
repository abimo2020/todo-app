package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
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

	setupLogger(cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect rabbitmq: %v", err))
		os.Exit(1)
	}
	defer conn.Close()

	consumer := todocreated.New(cfg.RabbitMQURL)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		if err := consumer.Run(context.Background()); err != nil {
			slog.Error(fmt.Sprintf("Consumer stopped: %v", err))
			stop()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}

	go func() {
		slog.Info(fmt.Sprintf("Server berjalan di :%s", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("Stop signal retrieved, starting shutdown..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error(fmt.Sprintf("failed to shutdown http server: %v", err))
	}

	select {
	case <-consumerDone:
		slog.Info("Consumer done")
	case <-shutdownCtx.Done():
		slog.Info("Timeout waiting consumer")
	}
	slog.Info("shutdown finish.")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "notif-svc",
	})
}

func setupLogger(format string) {
	opts := slog.HandlerOptions{Level: slog.LevelInfo}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &opts)
	}

	slog.SetDefault(slog.New(handler))
}
