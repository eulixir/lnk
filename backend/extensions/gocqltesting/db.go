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

const (
	connectTimeout = 30 * time.Second
	timeout        = 10 * time.Second
)

var nonAlphaRegex = regexp.MustCompile(`\W`)

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

	dropQuery := "DROP KEYSPACE IF EXISTS " + dbName
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
		_, _ = fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	cluster.Keyspace = dbName
	cluster.ConnectTimeout = connectTimeout
	cluster.Timeout = timeout
	cluster.Consistency = cassandra.Quorum

	testSession, err := cluster.CreateSession()
	if err != nil {
		_ = session.Query(dropQuery).Exec()
		return nil, fmt.Errorf("failed to create session with keyspace: %w", err)
	}

	t.Cleanup(func() {
		testSession.Close()

		_ = session.Query("DROP KEYSPACE IF EXISTS " + dbName).Exec()
	})

	return testSession, nil
}

// copySchemaFromTemplate copies all tables and indexes from the template keyspace to the target keyspace
// by querying system_schema tables. This provides fast schema replication without running migrations.
func copySchemaFromTemplate(session *cassandra.Session, targetKeyspace string) error {
	tableNames, err := getTableNames(session)
	if err != nil {
		return err
	}

	for _, tableName := range tableNames {
		err := copyTable(session, targetKeyspace, tableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func getTableNames(session *cassandra.Session) ([]string, error) {
	iter := session.Query(`
		SELECT table_name 
		FROM system_schema.tables 
		WHERE keyspace_name = ?
	`, _templateKeyspace).Iter()

	var (
		tableNames []string
		tableName  string
	)

	for iter.Scan(&tableName) {
		if tableName == "schema_migrations" {
			continue
		}

		tableNames = append(tableNames, tableName)
	}

	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tableNames, nil
}

type columnInfo struct {
	name     string
	typ      string
	kind     string
	position int
}

func copyTable(session *cassandra.Session, targetKeyspace, tableName string) error {
	columnInfos, err := getColumnInfos(session, tableName)
	if err != nil {
		return err
	}

	if err := createTable(session, targetKeyspace, tableName, columnInfos); err != nil {
		return err
	}

	return copyIndexes(session, targetKeyspace, tableName)
}

func getColumnInfos(session *cassandra.Session, tableName string) ([]columnInfo, error) {
	colIter := session.Query(`
		SELECT column_name, type, kind, position
		FROM system_schema.columns
		WHERE keyspace_name = ? AND table_name = ?
	`, _templateKeyspace, tableName).Iter()

	var (
		columnInfos                        []columnInfo
		columnName, columnType, columnKind string
		position                           int
	)

	for colIter.Scan(&columnName, &columnType, &columnKind, &position) {
		columnInfos = append(columnInfos, columnInfo{
			name:     columnName,
			typ:      columnType,
			kind:     columnKind,
			position: position,
		})
	}

	err := colIter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	sort.Slice(columnInfos, func(i, j int) bool {
		return columnInfos[i].position < columnInfos[j].position
	})

	return columnInfos, nil
}

func createTable(session *cassandra.Session, targetKeyspace, tableName string, columnInfos []columnInfo) error {
	columns := make([]string, 0, len(columnInfos))

	var (
		partitionKeys  []string
		clusteringKeys []string
	)

	for _, col := range columnInfos {
		columns = append(columns, fmt.Sprintf("%s %s", col.name, col.typ))
		switch col.kind {
		case "partition_key":
			partitionKeys = append(partitionKeys, col.name)
		case "clustering":
			clusteringKeys = append(clusteringKeys, col.name)
		}
	}

	createStmt := buildCreateTableStatement(targetKeyspace, tableName, columns, partitionKeys, clusteringKeys)

	err := session.Query(createStmt).Exec()
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return nil
}

func buildCreateTableStatement(targetKeyspace, tableName string, columns, partitionKeys, clusteringKeys []string) string {
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

	return createStmt
}

func copyIndexes(session *cassandra.Session, targetKeyspace, tableName string) error {
	idxIter := session.Query(`
		SELECT index_name, kind, options
		FROM system_schema.indexes
		WHERE keyspace_name = ? AND table_name = ?
	`, _templateKeyspace, tableName).Iter()

	var indexName, kind string
	for idxIter.Scan(&indexName, &kind, new(string)) {
		if kind == "KEYS" {
			continue
		}

		columnName := extractColumnFromIndexName(indexName, tableName)
		if columnName != "" {
			createIndexStmt := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s.%s (%s)",
				indexName, targetKeyspace, tableName, columnName)

			err := session.Query(createIndexStmt).Exec()
			if err != nil {
				_ = err
			}
		}
	}

	err := idxIter.Close()
	if err != nil {
		_ = err
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
