package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	"github.com/abimo2020/todo-svc/config"
	"github.com/abimo2020/todo-svc/internal/gateway/notif"
	"github.com/abimo2020/todo-svc/internal/handler/todo"
	"github.com/abimo2020/todo-svc/internal/middleware"
	todoRepo "github.com/abimo2020/todo-svc/internal/repository/todo"
	todoSvc "github.com/abimo2020/todo-svc/internal/service/todo"
	"github.com/abimo2020/todo-svc/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.Setup(cfg.LogFormat)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("ping failure")
		os.Exit(1)
	}

	notifClient := notif.NewRabbitMQ(cfg.RabbitMQURL)
	defer notifClient.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	todoRepo := todoRepo.New(db)
	todoSvc := todoSvc.New(todoRepo, notifClient)
	todoHandler := todo.New(todoSvc)

	todoHandler.RegisterRoutes(mux)

	srv := http.Server{Addr: ":" + cfg.Port, Handler: middleware.Logging(mux)}

	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Port).Msg("todo-svc running")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("receive stop signal, starting shutdown")
	case err := <-serverErr:
		log.Info().Err(err).Msg("http server error")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Info().Err(err).Msg("failed to shutdown server")
	}
	log.Info().Msg("success to shutdown")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"service": "todo-svc",
	})
}

// func setupLogger(format string) {
// 	opts := slog.HandlerOptions{
// 		Level: slog.LevelInfo,
// 	}

// 	var handler slog.Handler
// 	if format == "json" {
// 		handler = slog.NewJSONHandler(os.Stdout, &opts)
// 	} else {
// 		handler = slog.NewTextHandler(os.Stdout, &opts)
// 	}

// 	slog.SetDefault(slog.New(handler))
// }
