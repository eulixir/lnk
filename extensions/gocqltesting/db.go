package gocqltesting

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	cassandra "github.com/apache/cassandra-gocql-driver/v2"
)

var nonAlphaRegex = regexp.MustCompile(`[\W]`)

// NewDB creates a new isolated test keyspace by copying the schema from the template keyspace.
// This is much faster than running migrations for each test. The keyspace is automatically
// cleaned up when the test completes.
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

	if err := copySchemaFromTemplate(session, dbName); err != nil {
		_ = session.Query(dropQuery).Exec()
		return nil, fmt.Errorf("failed to copy schema from template: %w", err)
	}

	port := 9042
	if dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	cluster.Keyspace = dbName
	cluster.ConnectTimeout = 30 * time.Second
	cluster.Timeout = 10 * time.Second
	cluster.Consistency = cassandra.Quorum

	testSession, err := cluster.CreateSession()
	if err != nil {
		_ = session.Query(dropQuery).Exec()
		return nil, fmt.Errorf("failed to create session with keyspace: %w", err)
	}

	t.Cleanup(func() {
		testSession.Close()
		_ = session.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", dbName)).Exec()
	})

	return testSession, nil
}

// copySchemaFromTemplate copies all tables and indexes from the template keyspace to the target keyspace
// by querying system_schema tables. This provides fast schema replication without running migrations.
func copySchemaFromTemplate(session *cassandra.Session, targetKeyspace string) error {
	iter := session.Query(`
		SELECT table_name 
		FROM system_schema.tables 
		WHERE keyspace_name = ?
	`, _templateKeyspace).Iter()

	var tableNames []string
	var tableName string
	for iter.Scan(&tableName) {
		if tableName == "schema_migrations" {
			continue
		}
		tableNames = append(tableNames, tableName)
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	for _, tableName := range tableNames {
		colIter := session.Query(`
			SELECT column_name, type, kind, position
			FROM system_schema.columns
			WHERE keyspace_name = ? AND table_name = ?
		`, _templateKeyspace, tableName).Iter()

		type columnInfo struct {
			name     string
			typ      string
			kind     string
			position int
		}

		var columnInfos []columnInfo
		var columnName, columnType, columnKind string
		var position int

		for colIter.Scan(&columnName, &columnType, &columnKind, &position) {
			columnInfos = append(columnInfos, columnInfo{
				name:     columnName,
				typ:      columnType,
				kind:     columnKind,
				position: position,
			})
		}
		if err := colIter.Close(); err != nil {
			return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}

		sort.Slice(columnInfos, func(i, j int) bool {
			return columnInfos[i].position < columnInfos[j].position
		})

		var columns []string
		var partitionKeys []string
		var clusteringKeys []string

		for _, col := range columnInfos {
			columns = append(columns, fmt.Sprintf("%s %s", col.name, col.typ))
			switch col.kind {
			case "partition_key":
				partitionKeys = append(partitionKeys, col.name)
			case "clustering":
				clusteringKeys = append(clusteringKeys, col.name)
			}
		}

		createStmt := fmt.Sprintf("CREATE TABLE %s.%s (", targetKeyspace, tableName)
		createStmt += strings.Join(columns, ", ")
		createStmt += ", PRIMARY KEY ("

		if len(partitionKeys) > 0 {
			if len(partitionKeys) == 1 {
				createStmt += partitionKeys[0]
			} else {
				createStmt += "(" + strings.Join(partitionKeys, ", ") + ")"
			}
			if len(clusteringKeys) > 0 {
				createStmt += ", " + strings.Join(clusteringKeys, ", ")
			}
		}
		createStmt += "))"

		if err := session.Query(createStmt).Exec(); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}

		idxIter := session.Query(`
			SELECT index_name, kind, options
			FROM system_schema.indexes
			WHERE keyspace_name = ? AND table_name = ?
		`, _templateKeyspace, tableName).Iter()

		var indexName, kind, options string
		for idxIter.Scan(&indexName, &kind, &options) {
			if kind == "KEYS" {
				continue
			}

			columnName := extractColumnFromIndexName(indexName, tableName)
			if columnName != "" {
				createIndexStmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s.%s (%s)",
					indexName, targetKeyspace, tableName, columnName)
				if err := session.Query(createIndexStmt).Exec(); err != nil {
					_ = err
				}
			}
		}
		if err := idxIter.Close(); err != nil {
			_ = err
		}
	}

	return nil
}

// extractColumnFromIndexName extracts the column name from an index name using naming patterns.
// Pattern: table_column_idx -> column
func extractColumnFromIndexName(indexName, tableName string) string {
	prefix := tableName + "_"
	if strings.HasPrefix(indexName, prefix) {
		suffix := strings.TrimPrefix(indexName, prefix)
		suffix = strings.TrimSuffix(suffix, "_idx")
		suffix = strings.TrimSuffix(suffix, "_index")
		return suffix
	}
	return ""
}
