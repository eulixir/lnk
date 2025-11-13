package gcqltesting

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	cassandra "github.com/apache/cassandra-gocql-driver/v2"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
)

var nonAlphaRegex = regexp.MustCompile(`[\W]`)

func NewDB(t *testing.T, dbName string) (*cassandra.Session, error) {
	t.Helper()

	session := concurrentSession

	if session == nil {
		return nil, errors.New("session is nil - ensure StartDockerContainer has been called")
	}

	if dbName == "" {
		return nil, errors.New("dbName cannot be an empty string")
	}
	dbName = nonAlphaRegex.ReplaceAllString(strings.ToLower(dbName), "_")

	dropQuery := fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", dbName)
	if err := session.Query(dropQuery).Exec(); err != nil {
		return nil, fmt.Errorf("failed to drop existing keyspace: %w", err)
	}

	createKeyspaceQuery := fmt.Sprintf(
		"CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		dbName,
	)
	if err := session.Query(createKeyspaceQuery).Exec(); err != nil {
		return nil, fmt.Errorf("failed to create keyspace: %w", err)
	}

	t.Cleanup(func() {
		session.Close()
		_ = session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", dbName)).Exec()
	})

	return session, nil
}
