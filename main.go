package main

import (
	"context"
	"fmt"
	"log"

	"lnk/extensions/config"
	"lnk/extensions/logger"
	"lnk/gateways/gocql"

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
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer session.Close()
	log.Println("Database connection established successfully")
}
func setupDatabase(ctx context.Context, config *config.Config, logger *zap.Logger) (*gocql.Session, error) {
	session, err := gocql.SetupDatabase(ctx, &config.Gocql, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return session, nil
}
