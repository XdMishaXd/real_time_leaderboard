package main

import (
	"auth_service/internal/auth"
	"auth_service/internal/config"
	"auth_service/internal/http_server/handlers/login"
	"auth_service/internal/http_server/handlers/logout"
	"auth_service/internal/http_server/handlers/refresh"
	register "auth_service/internal/http_server/handlers/register"
	"auth_service/internal/http_server/handlers/verify"
	"auth_service/internal/rabbitmq"
	"auth_service/internal/storage/postgres"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad("./config/config.yaml")

	log := setupLogger(cfg.Env)

	log.Info("starting auth service", slog.String("env", cfg.Env))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("Shutdown signal received")
		cancel()
	}()

	storage, err := postgres.New(ctx, cfg)
	if err != nil {
		log.Error("failed to connect postgres", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer storage.Close()

	msgBroker, err := rabbitmq.New(cfg.RabbitMQ.URL, cfg.RabbitMQ.QueueName)
	if err != nil {
		log.Error("failed to connect rabbitmq", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer msgBroker.Close()

	authService := auth.New(log, storage, storage, storage, cfg.Tokens.AccessTokenTTL, cfg.Tokens.RefreshTokenTTL)

	router := setupRouter(
		ctx,
		log,
		*authService,
		*msgBroker,
		cfg.Tokens.VerificationTokenTTL,
		cfg.Tokens.VerificationTokenSecret,
		cfg.HTTPServer.Address,
	)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		log.Info("HTTP server is running")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", slog.String("err", err.Error()))
			cancel()
		}
	}()

	<-ctx.Done()

	log.Info("Shutting down HTTP server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server shutdown error", slog.String("err", err.Error()))
	} else {
		log.Info("Server stopped gracefully")
	}

	log.Info("Main service stopped")
}

func setupRouter(
	ctx context.Context,
	log *slog.Logger,
	authService auth.Auth,
	msgBroker rabbitmq.RabbitMQClient,
	verificationTokenTTL time.Duration,
	verificationTokenSecret string,
	address string,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/register",
		register.New(ctx, log, authService, &msgBroker, verificationTokenTTL, verificationTokenSecret, address),
	)
	r.Post("/login",
		login.New(ctx, log, authService),
	)
	r.Post("/refresh",
		refresh.New(ctx, log, authService),
	)
	r.Post("/logout",
		logout.New(ctx, log, authService),
	)
	r.Get("/verify",
		verify.New(ctx, log, authService, verificationTokenSecret),
	)

	return r
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
