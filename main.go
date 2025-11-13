package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"lnk/extensions/config"
	"lnk/extensions/logger"
	"lnk/gateways/gocql"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(config.Logger)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("Starting application")

	session, err := setupDatabase(context.Background(), config, logger)
	if err != nil {
		logger.Error("Failed to setup database", zap.Error(err))
	}
	defer session.Close()
	logger.Info("Database connection established successfully")
}

func startServer(ctx context.Context, config *config.Config, logger *zap.Logger) {
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "OK"})
	})
	router.Run(config.Port)
}

func setupDatabase(ctx context.Context, config *config.Config, logger *zap.Logger) (*gocql.Session, error) {
	session, err := gocql.SetupDatabase(ctx, &config.Gocql, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return session, nil
}
