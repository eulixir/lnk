package main

import (
	"fmt"
	"log"

	"lnk/extensions/config"
	"lnk/extensions/logger"
	gocqlPackage "lnk/gateways/gocql"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"go.uber.org/zap"
)

func main() {
	cfg, logger := setupConfigAndLogger()

	session, err := setupDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	defer session.Close()

	logger.Info("Migrations completed successfully")
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
	session, err := gocqlPackage.SetupDatabase(&cfg.Gocql, logger, cfg.Gocql.AutoMigrate)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return session, nil
}
