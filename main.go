package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lnk/extensions/config"
	"lnk/extensions/logger"
	"lnk/extensions/redis"
	"lnk/gateways/gocql"
	httpServer "lnk/gateways/http"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Starting application")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := gocql.SetupDatabase(ctx, &cfg.Gocql, logger)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	defer session.Close()

	redisClient, err := redis.SetupRedis(ctx, &cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to setup Redis", zap.Error(err))
	}
	defer redisClient.Close()

	server := httpServer.NewServer(logger, cfg.App.Port, cfg.App.GinMode)
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}

	logger.Info("Application stopped")
}
