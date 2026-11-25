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

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/abimo2020/notif-svc/config"
	"github.com/abimo2020/notif-svc/internal/consumer/todocreated"
	notifHandler "github.com/abimo2020/notif-svc/internal/handler/notification"
	"github.com/abimo2020/notif-svc/internal/middleware"
	notifRepo "github.com/abimo2020/notif-svc/internal/repository/notification"
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
	}
	defer conn.Close()

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create mongodb client")
	}

	defer func() {
		dc, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		mongoClient.Disconnect(dc)
	}()

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	if err := mongoClient.Ping(pingCtx, nil); err != nil {
		cancelPing()
		log.Fatal().Err(err).Msg("cannot call mongodb")
	}

	notifRepo := notifRepo.New(mongoClient.Database("notif_db"))

	if err := notifRepo.EnsureIndexes(ctx); err != nil {
		cancelPing()
		log.Fatal().Err(err).Msg("failed to create index")
	}
	cancelPing()

	consumer := todocreated.New(cfg.RabbitMQURL, notifRepo)
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		if err := consumer.Run(context.Background()); err != nil {
			log.Fatal().Err(err).Msg("Consumer stopped")
			stop()
		}
	}()

	router := gin.New()
	router.SetTrustedProxies(nil)
	router.Use(middleware.Logging(), gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "notif-svc"})
	})
	notifHandler.New(notifRepo).RegisterRoutes(router)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}

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
