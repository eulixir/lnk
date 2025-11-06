package gocql

import (
	"context"
	"fmt"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type Session struct {
	*gocql.Session
}

func SetupDatabase(ctx context.Context, config *Config) (*Session, error) {
	cluster := gocql.NewCluster(config.Host)
	cluster.Port = config.Port
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: config.Username,
		Password: config.Password,
	}
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	createKeyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		config.Keyspace,
	)
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create keyspace: %w", err)
	}

	session.Close()

	cluster.Keyspace = config.Keyspace
	session, err = cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session with keyspace: %w", err)
	}

	return &Session{Session: session}, nil
}
