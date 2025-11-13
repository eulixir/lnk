package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lnk/domain/entities/usecases"
	"lnk/extensions/config"
	"lnk/extensions/logger"
	"lnk/extensions/redis"
	"lnk/gateways/gocql"
	"lnk/gateways/gocql/repositories"
	httpServer "lnk/gateways/http"
	"lnk/gateways/http/handlers"
	"lnk/gateways/http/middleware"

	"github.com/gin-gonic/gin"
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

	repository := repositories.NewRepository(logger, session)
	useCase := usecases.NewUseCase(logger, repository)

	router := setupGinEngine(logger, cfg)

	httpHandlers := handlers.NewHttpHandlers(&handlers.HttpHandlers{
		Router:  router,
		Logger:  logger,
		Env:     cfg.App.ENV,
		UseCase: useCase,
	})
	_ = httpHandlers // handlers are registered to router during initialization

	server := httpServer.NewServer(logger, cfg.App.Port, router)
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

func setupGinEngine(logger *zap.Logger, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.App.GinMode)
	router := gin.New()
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS())

	return router
}
