package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/abimo2020/notif-svc/config"
	"github.com/abimo2020/notif-svc/internal/consumer/todocreated"
	"github.com/abimo2020/notif-svc/pkg/logger"
)

const (
	PORT = 8081
)

func main() {
	cfg := config.Load()

	logger.Setup(cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect rabbitmq")
		os.Exit(1)
	}
	defer conn.Close()

	consumer := todocreated.New(cfg.RabbitMQURL)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		if err := consumer.Run(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("Consumer stopped")
			stop()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: mux}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("notif-svc running")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("http server error")
			stop()
		}
	}()

	<-ctx.Done()
	log.Info().Msg("Stop signal retrieved, starting shutdown..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("failed to shutdown http server")
	}

	select {
	case <-consumerDone:
		log.Info().Msg("Consumer done")
	case <-shutdownCtx.Done():
		log.Info().Msg("Timeout waiting consumer")
	}
	log.Info().Msg("shutdown finish.")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "notif-svc",
	})
}

// func setupLogger(format string) {
// 	opts := slog.HandlerOptions{Level: slog.LevelInfo}

// 	var handler slog.Handler
// 	if format == "json" {
// 		handler = slog.NewJSONHandler(os.Stdout, &opts)
// 	} else {
// 		handler = slog.NewTextHandler(os.Stdout, &opts)
// 	}

// 	slog.SetDefault(slog.New(handler))
// }
