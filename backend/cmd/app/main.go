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
	cfg, appLogger := setupConfigAndLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownOTel, err := setupOTelSDK(ctx, cfg)
	if err != nil {
		appLogger.Fatal("Failed to setup OpenTelemetry", zap.Error(err))
	}
	defer func() {
		if err := shutdownOTel(ctx); err != nil {
			appLogger.Error("Failed to shutdown OpenTelemetry", zap.Error(err))
		}
	}()

	session := setupDatabase(cfg, appLogger)
	defer session.Close()

	redisClient := setupRedis(ctx, cfg, appLogger)

	defer func() {
		err := redisClient.Close()
		if err != nil {
			appLogger.Error("Failed to close Redis client", zap.Error(err))
		}
	}()

	initializeCounter(ctx, redisClient, cfg, appLogger)

	useCase := createUseCase(cfg, appLogger, session, redisClient)
	server := createAndStartServer(cfg, appLogger, useCase)

	shutdownServer(ctx, appLogger, server)
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

	appLogger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	appLogger.Info("Starting application")

	return cfg, appLogger
}

func setupDatabase(cfg *config.Config, appLogger *zap.Logger) *gocql.Session {
	session, err := gocqlPackage.SetupDatabase(&cfg.Gocql, appLogger, false)
	if err != nil {
		appLogger.Fatal("Failed to setup database", zap.Error(err))
	}

	return session
}

func setupRedis(ctx context.Context, cfg *config.Config, appLogger *zap.Logger) *redis.Client {
	redisClient, err := redisPackage.SetupRedis(ctx, &cfg.Redis, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to setup Redis", zap.Error(err))
	}

	return redisClient
}

func initializeCounter(ctx context.Context, redisClient *redis.Client, cfg *config.Config, appLogger *zap.Logger) {
	setInitialCounter, err := redisPackage.SetInitialCounterValue(ctx, redisClient, &cfg.Redis, appLogger)
	if err != nil && !setInitialCounter {
		appLogger.Fatal("Failed to set initial counter", zap.Error(err))
	}
}

func createUseCase(cfg *config.Config, appLogger *zap.Logger, session *gocql.Session, redisClient *redis.Client) *usecases.UseCase {
	repository := repositories.NewRepository(appLogger, session)
	redisAdapter := redisPackage.NewRedisAdapter(redisClient)

	return usecases.NewUseCase(usecases.NewUseCaseParams{
		Logger:     appLogger,
		Repository: repository,
		Redis:      redisAdapter,
		Salt:       cfg.App.Base62Salt,
		CounterKey: cfg.Redis.CounterKey,
	})
}

func createAndStartServer(cfg *config.Config, appLogger *zap.Logger, useCase *usecases.UseCase) *httpServer.Server {
	httpHandlers := handlers.NewHandlers(appLogger, useCase)

	router := httpServer.NewRouter(httpServer.RouterConfig{
		Logger:   appLogger,
		GinMode:  cfg.App.GinMode,
		Env:      cfg.App.ENV,
		Handlers: httpHandlers,
	})

	server := httpServer.NewServer(appLogger, cfg.App.Port, router)

	err := server.Start()
	if err != nil {
		appLogger.Fatal("Failed to start HTTP server", zap.Error(err))
	}

	return server
}

func shutdownServer(ctx context.Context, appLogger *zap.Logger, server *httpServer.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	appLogger.Info("Received shutdown signal")

	const shutdownTimeout = 10 * time.Second

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		appLogger.Error("Error during server shutdown", zap.Error(err))
	}

	appLogger.Info("Application stopped")
}
