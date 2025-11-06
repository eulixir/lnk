package main

import (
	"context"
	"fmt"
	"log"

	"lnk/extensions/config"
	"lnk/gateways/gocql"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	session, err := setupDatabase(context.Background(), config)
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer session.Close()
	log.Println("Database connection established successfully")
}
func setupDatabase(ctx context.Context, config *config.Config) (*gocql.Session, error) {
	session, err := gocql.SetupDatabase(ctx, &config.Gocql)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}
	return session, nil
}
