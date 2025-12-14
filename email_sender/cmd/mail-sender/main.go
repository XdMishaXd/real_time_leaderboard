package main

import (
	"context"
	"email_sender/internal/config"
	sl "email_sender/internal/lib/logger"
	mailer "email_sender/internal/mail-sender"
	"email_sender/internal/models"
	"email_sender/internal/rabbitmq"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("Starting email_sender", slog.String("env", cfg.Env))

	startServer(ctx, cfg, log)
}

func startServer(ctx context.Context, cfg *config.Config, log *slog.Logger) {
	r, err := rabbitmq.New(cfg.RabbitMQURL)
	if err != nil {
		log.Error("failed to init rabbitmq", sl.Err(err))
		return
	}
	defer r.Close()

	m := &mailer.Mailer{
		Host:     cfg.Email.Host,
		Port:     cfg.Email.Port,
		Username: cfg.Email.Username,
		Password: cfg.Email.Password,
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		err := r.StartReading(ctx, cfg.QueueName, func(msg []byte) {
			var emailMsg models.EmailMessage
			if err := json.Unmarshal(msg, &emailMsg); err != nil {
				log.Error("failed to unmarshal message", sl.Err(err))
				return
			}

			err := m.Send(emailMsg.Email,
				emailMsg.Subject,
				"http://localhost"+emailMsg.MessageText,
			)
			if err != nil {
				log.Error("failed to send message", sl.Err(err))
				return
			}

			log.Info("message sent successfully")
		})
		if err != nil {
			log.Error("failed to start reading", sl.Err(err))
			return
		}
	}()

	log.Info("consumer successfully started")

	select {
	case <-ctx.Done():
		log.Info("shutting down consumer...")
	case <-done:
		log.Info("consumer finished the work")
	}

	log.Info("service gracefully stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
