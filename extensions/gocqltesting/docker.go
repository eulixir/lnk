package gocqltesting

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	cassandra "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	migrationHTTPS "github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	_reusableContainerName     = "gocql-testing-container"
	_defaultGocqlDockerVersion = "latest"
	_templateKeyspace          = "template_keyspace"
	retryIntervalMs            = 500
)

var (
	dbPort               string
	setupMutex           sync.Mutex
	containerInitialized bool
	concurrentSession    *cassandra.Session
	templateReady        bool
)

type Migrations struct {
	FS fs.FS
}

type DockerContainerConfig struct {
	ReuseContainer bool
	Version        string
	ContainerName  string
	Migrations     *Migrations
}

// StartDockerContainer starts a Cassandra Docker container and sets up a template keyspace with migrations.
// It returns a teardown function. The template keyspace is created once and reused across tests for fast schema copying.
func StartDockerContainer(cfg DockerContainerConfig) (teardownFn func(), err error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if containerInitialized && concurrentSession != nil {
		return func() {}, nil
	}

	dockerPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

	dockerPool.MaxWait = 180 * time.Second

	if err = dockerPool.Client.Ping(); err != nil {
		return nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

	if cfg.Version == "" {
		cfg.Version = _defaultGocqlDockerVersion
	}

	dockerResource, err := getDockerGocqlResource(dockerPool, cfg)
	if err != nil {
		return nil, fmt.Errorf(`failed to initialize gocql docker resource: %w`, err)
	}

	dbPort = dockerResource.GetPort("9042/tcp")

	startTime := time.Now()
	maxWait := 180 * time.Second
	backoff := 2 * time.Second

	for {
		if time.Since(startTime) > maxWait {
			return nil, fmt.Errorf("cassandra not ready after %v: timeout exceeded", maxWait)
		}

		err = pingCassandraFn(dbPort)()
		if err == nil {
			break
		}

		fmt.Printf("Cassandra not ready yet (attempt after %v): %v\n", time.Since(startTime), err)
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}

	time.Sleep(3 * time.Second)

	concurrentSession, err = newDB(getCassandraConnString(dbPort, "master"))
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	containerInitialized = true

	if !templateReady {
		if err := setupTemplateDatabase(concurrentSession, cfg.Migrations.FS); err != nil {
			concurrentSession.Close()
			return nil, err
		}
		templateReady = true
	}

	teardownFn = func() {
		if !cfg.ReuseContainer {
			if concurrentSession != nil {
				concurrentSession.Close()
				concurrentSession = nil
			}
			_ = dockerResource.Close()
			containerInitialized = false
			templateReady = false
		}
	}

	return teardownFn, nil
}

// setupTemplateDatabase creates a template keyspace and runs migrations on it once.
// This template is used to quickly copy schema to test keyspaces without running migrations each time.
func setupTemplateDatabase(conn *cassandra.Session, migrationsFs fs.FS) error {
	dbKeyspace := _templateKeyspace

	var dbCount int
	err := conn.Query("SELECT COUNT(*) FROM system_schema.keyspaces WHERE keyspace_name = ?", dbKeyspace).Scan(&dbCount)
	if err != nil {
		return fmt.Errorf("error checking template keyspace: %w", err)
	}

	if dbCount > 0 {
		return nil
	}

	fmt.Println("Creating template keyspace and running migrations")

	createQuery := fmt.Sprintf(
		"CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		dbKeyspace,
	)
	if err := conn.Query(createQuery).Exec(); err != nil {
		return fmt.Errorf("error creating template keyspace: %w", err)
	}

	time.Sleep(time.Millisecond * retryIntervalMs)

	err = runMigrations(getCassandraConnString(dbPort, dbKeyspace), migrationsFs)
	if err != nil {
		return fmt.Errorf("failed to run migrations on template: %w", err)
	}

	time.Sleep(time.Millisecond * retryIntervalMs)

	return nil
}

// getDockerGocqlResource gets or creates a Cassandra Docker container resource.
func getDockerGocqlResource(dockerPool *dockertest.Pool, cfg DockerContainerConfig) (*dockertest.Resource, error) {
	var containerName string
	if cfg.ReuseContainer {
		containerName = _reusableContainerName
		if cfg.ContainerName != "" {
			containerName = cfg.ContainerName
		}
		container, _ := dockerPool.Client.InspectContainer(containerName)
		if container != nil && container.State.Running {
			resource := &dockertest.Resource{Container: container}
			return resource, nil
		}
		if container != nil && !container.State.Running {
			_ = dockerPool.RemoveContainerByName(containerName)
		}
	}

	resource, err := dockerPool.RunWithOptions(&dockertest.RunOptions{
		Name:       containerName,
		Repository: "cassandra",
		Tag:        cfg.Version,
		Env: []string{
			"CASSANDRA_CLUSTER_NAME=lnk-cluster",
			"CASSANDRA_DC=datacenter1",
			"CASSANDRA_RACK=rack1",
		},
	}, func(c *docker.HostConfig) {
		c.AutoRemove = true
		c.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err == nil {
		return resource, nil
	}

	return nil, fmt.Errorf(`RunWithOptions failed: %w`, err)
}

// pingCassandraFn returns a function that tests if Cassandra is ready by attempting a connection and query.
func pingCassandraFn(port string) func() error {
	return func() error {
		portInt := 9042
		if port != "" {
			_, _ = fmt.Sscanf(port, "%d", &portInt)
		}

		cluster := cassandra.NewCluster("localhost")
		cluster.Port = portInt
		cluster.ConnectTimeout = 5 * time.Second
		cluster.Timeout = 3 * time.Second
		cluster.Consistency = cassandra.One
		cluster.DisableInitialHostLookup = true
		cluster.NumConns = 1

		conn, err := cluster.CreateSession()
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		defer conn.Close()
		err = conn.Query("SELECT now() FROM system.local").Exec()
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		return nil
	}
}

// getCassandraConnString returns a Cassandra connection string for migrations.
// Format: cassandra://host:port/keyspace?x-multi-statement=true
func getCassandraConnString(port, keyspace string) string {
	return fmt.Sprintf("cassandra://localhost:%s/%s?x-multi-statement=true", port, keyspace)
}

// newDB creates a new Cassandra session using the dbPort package variable.
func newDB(_ string) (*cassandra.Session, error) {
	port := 9042
	if dbPort != "" {
		_, _ = fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	cluster.ConnectTimeout = 30 * time.Second
	cluster.Timeout = 10 * time.Second
	cluster.Consistency = cassandra.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

// runMigrations runs database migrations from the provided filesystem on the specified connection.
func runMigrations(connectionString string, migrationsFs fs.FS) error {
	sourceDriver, err := migrationHTTPS.New(http.FS(migrationsFs), ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("httpfs", sourceDriver, connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
