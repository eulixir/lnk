package gocql

import (
	"errors"
	"fmt"
	"time"

	"lnk/gateways/gocql/migrations"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

func SetupDatabase(config *Config, logger *zap.Logger) (*gocql.Session, error) {
	cluster := gocql.NewCluster(config.Host)
	cluster.Port = config.Port
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	cluster.Consistency = gocql.Quorum
	cluster.ConnectTimeout = 30 * time.Second

	session, err := createSessionWithRetry(cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	createKeyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		config.Keyspace,
	)
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		logger.Warn("Failed to create keyspace (may already exist)", zap.Error(err))
	}

	cluster.Keyspace = config.Keyspace
	sessionWithKeyspace, err := cluster.CreateSession()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create session with keyspace: %w", err)
	}
	session.Close()
	session = sessionWithKeyspace

	if config.AutoMigrate {
		if err := runMigrations(config, logger); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	logger.Info("Database connection established successfully")

	return session, nil
}

func createSessionWithRetry(cluster *gocql.ClusterConfig) (*gocql.Session, error) {
	const (
		maxAttempts = 10
		backoff     = 3 * time.Second
	)

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		session, err := cluster.CreateSession()
		if err == nil {
			return session, nil
		}
		lastErr = err
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("failed to create session after %d attempts: %w", maxAttempts, lastErr)
}

func runMigrations(config *Config, logger *zap.Logger) error {
	sourceDriver, err := iofs.New(migrations.MigrationsFS, ".")
	if err != nil {
		logger.Error("failed to load embedded migrations", zap.Error(err))
		return fmt.Errorf("failed to load embedded migrations: %w", err)
	}

	migrationURL := fmt.Sprintf("cassandra://%s:%d/%s?x-multi-statement=true", config.Host, config.Port, config.Keyspace)

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrationURL)
	if err != nil {
		logger.Error("failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	logger.Info("Running migrations")
	defer func() {
		_, err := m.Close()
		if err != nil {
			logger.Warn("Failed to close migration instance", zap.Error(err))
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No migration changes to apply")
			return nil
		}
		return err
	}
	return nil
}
