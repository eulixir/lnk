package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"lnk/domain/entities/usecases"
	"lnk/extensions/config"
	"lnk/extensions/logger"
	redisPackage "lnk/extensions/redis"
	gocqlPackage "lnk/gateways/gocql"
	"lnk/gateways/gocql/repositories"
	httpServer "lnk/gateways/http"
	"lnk/gateways/http/handlers"
)

func main() {
	cfg, logger := setupConfigAndLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session := setupDatabase(cfg, logger)
	defer session.Close()

	redisClient := setupRedis(ctx, cfg, logger)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis client", zap.Error(err))
		}
	}()

	initializeCounter(ctx, redisClient, cfg, logger)

	useCase := createUseCase(cfg, logger, session, redisClient)
	server := createAndStartServer(cfg, logger, useCase)

	shutdownServer(ctx, logger, server)
}

func setupConfigAndLogger() (*config.Config, *zap.Logger) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Starting application")
	return cfg, logger
}

func setupDatabase(cfg *config.Config, logger *zap.Logger) *gocql.Session {
	session, err := gocqlPackage.SetupDatabase(&cfg.Gocql, logger)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	return session
}

func setupRedis(ctx context.Context, cfg *config.Config, logger *zap.Logger) *redis.Client {
	redisClient, err := redisPackage.SetupRedis(ctx, &cfg.Redis, logger)
	if err != nil {
		logger.Fatal("Failed to setup Redis", zap.Error(err))
	}
	return redisClient
}

func initializeCounter(ctx context.Context, redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) {
	setInitialCounter, err := redisPackage.SetInitialCounterValue(ctx, redisClient, &cfg.Redis, logger)
	if err != nil && !setInitialCounter {
		logger.Fatal("Failed to set initial counter", zap.Error(err))
	}
}

func createUseCase(cfg *config.Config, logger *zap.Logger, session *gocql.Session, redisClient *redis.Client) *usecases.UseCase {
	repository := repositories.NewRepository(logger, session)
	redisAdapter := redisPackage.NewRedisAdapter(redisClient)

	return usecases.NewUseCase(usecases.NewUseCaseParams{
		Logger:     logger,
		Repository: repository,
		Redis:      redisAdapter,
		Salt:       cfg.App.Base62Salt,
		CounterKey: cfg.Redis.CounterKey,
	})
}

func createAndStartServer(cfg *config.Config, logger *zap.Logger, useCase *usecases.UseCase) *httpServer.Server {
	httpHandlers := handlers.NewHandlers(logger, useCase)

	router := httpServer.NewRouter(httpServer.RouterConfig{
		Logger:   logger,
		GinMode:  cfg.App.GinMode,
		Env:      cfg.App.ENV,
		Handlers: httpHandlers,
	})

	server := httpServer.NewServer(logger, cfg.App.Port, router)
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}

	return server
}

func shutdownServer(ctx context.Context, logger *zap.Logger, server *httpServer.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal")

	const shutdownTimeout = 10 * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}

	logger.Info("Application stopped")
}
