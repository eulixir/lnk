package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lnk/domain/entities/usecases"
	"lnk/extensions/config"
	"lnk/extensions/logger"
	"lnk/extensions/opentelemetry"
	redisPackage "lnk/extensions/redis"
	gocqlPackage "lnk/gateways/gocql"
	"lnk/gateways/gocql/repositories"
	httpServer "lnk/gateways/http"
	"lnk/gateways/http/handlers"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg, logger := setupConfigAndLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownOTel, err := setupOTelSDK(ctx, cfg)
	if err != nil {
		logger.Fatal("Failed to setup OpenTelemetry", zap.Error(err))
	}
	defer func() {
		if err := shutdownOTel(ctx); err != nil {
			logger.Error("Failed to shutdown OpenTelemetry", zap.Error(err))
		}
	}()

	session, err := setupDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer session.Close()

	redisClient, err := setupRedis(ctx, cfg, logger)
	if err != nil {
		log.Fatalf("Failed to setup Redis: %v", err)
	}

	defer func() {
		err := redisClient.Close()
		if err != nil {
			logger.Error("Failed to close Redis client", zap.Error(err))
		}
	}()

	setInitialCounter, err := initializeCounter(ctx, redisClient, cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize counter: %v", err)
	}
	if !setInitialCounter {
		log.Fatalf("Failed to set initial counter")
	}

	useCase := createUseCase(cfg, logger, session, redisClient)
	server, err := createAndStartServer(cfg, logger, useCase)
	if err != nil {
		log.Fatalf("Failed to create and start server: %v", err)
	}

	shutdownServer(ctx, logger, server)
}

func setupOTelSDK(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	shutdown, err := opentelemetry.SetupOTelSDK(ctx, &cfg.OTel)
	if err != nil {
		return nil, fmt.Errorf("failed to setup OpenTelemetry SDK: %w", err)
	}
	return shutdown, nil
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

func setupDatabase(cfg *config.Config, logger *zap.Logger) (*gocql.Session, error) {
	session, err := gocqlPackage.SetupDatabase(&cfg.Gocql, logger, false)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return session, nil
}

func setupRedis(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	redisClient, err := redisPackage.SetupRedis(ctx, &cfg.Redis, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Redis: %w", err)
	}

	return redisClient, nil
}

func initializeCounter(ctx context.Context, redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) (bool, error) {
	setInitialCounter, err := redisPackage.SetInitialCounterValue(ctx, redisClient, &cfg.Redis, logger)
	if err != nil && !setInitialCounter {
		return false, fmt.Errorf("failed to set initial counter: %w", err)
	}

	return setInitialCounter, nil
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

func createAndStartServer(cfg *config.Config, logger *zap.Logger, useCase *usecases.UseCase) (*httpServer.Server, error) {
	httpHandlers := handlers.NewHandlers(logger, useCase)

	router := httpServer.NewRouter(httpServer.RouterConfig{
		Logger:   logger,
		GinMode:  cfg.App.GinMode,
		Env:      cfg.App.ENV,
		Handlers: httpHandlers,
	})

	server := httpServer.NewServer(logger, cfg.App.Port, router)

	err := server.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return server, nil
}

func shutdownServer(ctx context.Context, logger *zap.Logger, server *httpServer.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal")

	const shutdownTimeout = 10 * time.Second

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}

	logger.Info("Application stopped")
}
