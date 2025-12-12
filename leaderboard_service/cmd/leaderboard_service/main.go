package main

import (
	"context"
	"leaderboard_service/internal/config"
	gametop "leaderboard_service/internal/http-server/handlers/game-top"
	myresult "leaderboard_service/internal/http-server/handlers/my-result"
	"leaderboard_service/internal/http-server/handlers/submit"
	"leaderboard_service/internal/lib/jwt"
	"leaderboard_service/internal/storage/redis"
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

	log.Info("starting leaderboard service", slog.String("env", cfg.Env))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("Shutdown signal received")
		cancel()
	}()

	redisClient, err := redis.New(ctx, cfg.Redis.Db, cfg.Redis.Addr)
	if err != nil {
		log.Error("failed to connect redis", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer redisClient.Close()

	jwtParser := jwt.New(cfg.JWTSecret)

	router := setupRouter(
		ctx,
		log,
		*redisClient,
		*jwtParser,
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

func setupRouter(ctx context.Context, log *slog.Logger, redisClient redis.RedisRepo, jwtParser jwt.JWTParser) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/submit", submit.New(ctx, log, &redisClient, jwtParser))
	r.Get("/{game}/top", gametop.New(ctx, log, &redisClient, jwtParser))
	r.Get("/{game}/me", myresult.New(ctx, log, &redisClient, jwtParser))

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
