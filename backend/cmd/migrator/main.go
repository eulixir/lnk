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
	cfg, appLogger := setupConfigAndLogger()

	session, err := setupDatabase(cfg, appLogger)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	defer session.Close()

	appLogger.Info("Migrations completed successfully")
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

func setupDatabase(cfg *config.Config, appLogger *zap.Logger) (*gocql.Session, error) {
	session, err := gocqlPackage.SetupDatabase(&cfg.Gocql, appLogger, cfg.Gocql.AutoMigrate)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return session, nil
}
