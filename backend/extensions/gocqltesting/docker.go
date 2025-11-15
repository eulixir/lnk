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
	maxWait                    = 180 * time.Second
	initialBackoff             = 2 * time.Second
	maxBackoff                 = 10 * time.Second
	postReadyDelay             = 3 * time.Second
	cassandraPort              = 9042
	pingConnectTimeout         = 5 * time.Second
	pingTimeout                = 3 * time.Second
	dockerConnectTimeout       = 30 * time.Second
	dockerTimeout              = 10 * time.Second
)

var (
	dbPort               string             //nolint:gochecknoglobals
	setupMutex           sync.Mutex         //nolint:gochecknoglobals
	containerInitialized bool               //nolint:gochecknoglobals
	concurrentSession    *cassandra.Session //nolint:gochecknoglobals
	templateReady        bool               //nolint:gochecknoglobals
)

type Migrations struct {
	FS fs.FS
}

type DockerContainerConfig struct {
	Migrations     *Migrations
	Version        string
	ContainerName  string
	ReuseContainer bool
}

// StartDockerContainer starts a Cassandra Docker container and sets up a template keyspace with migrations.
// It returns a teardown function. The template keyspace is created once and reused across tests for fast schema copying.
func StartDockerContainer(cfg DockerContainerConfig) (teardownFn func(), err error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if containerInitialized && concurrentSession != nil {
		return func() {}, nil
	}

	_, dockerResource, err := initializeDockerPool(cfg)
	if err != nil {
		return nil, err
	}

	if err := waitForCassandra(); err != nil {
		return nil, err
	}

	if err := initializeSession(cfg); err != nil {
		return nil, err
	}

	teardownFn = createTeardownFn(cfg, dockerResource)

	return teardownFn, nil
}

func initializeDockerPool(cfg DockerContainerConfig) (*dockertest.Pool, *dockertest.Resource, error) {
	dockerPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

	dockerPool.MaxWait = maxWait

	if err = dockerPool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

	if cfg.Version == "" {
		cfg.Version = _defaultGocqlDockerVersion
	}

	dockerResource, err := getDockerGocqlResource(dockerPool, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to initialize gocql docker resource: %w`, err)
	}

	dbPort = dockerResource.GetPort("9042/tcp")

	return dockerPool, dockerResource, nil
}

func waitForCassandra() error {
	startTime := time.Now()
	backoff := initialBackoff

	for {
		if time.Since(startTime) > maxWait {
			return fmt.Errorf("cassandra not ready after %v: timeout exceeded", maxWait)
		}

		err := pingCassandraFn(dbPort)()
		if err == nil {
			break
		}

		time.Sleep(backoff)

		if backoff < maxBackoff {
			backoff *= 2
		}
	}

	time.Sleep(postReadyDelay)

	return nil
}

func initializeSession(cfg DockerContainerConfig) error {
	var err error

	concurrentSession, err = newDB(getCassandraConnString(dbPort, "master"))
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	containerInitialized = true

	if !templateReady {
		err := setupTemplateDatabase(concurrentSession, cfg.Migrations.FS)
		if err != nil {
			concurrentSession.Close()
			return err
		}

		templateReady = true
	}

	return nil
}

func createTeardownFn(cfg DockerContainerConfig, dockerResource *dockertest.Resource) func() {
	return func() {
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

	// Creating template keyspace and running migrations

	createQuery := fmt.Sprintf(
		"CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		dbKeyspace,
	)

	execErr := conn.Query(createQuery).Exec()
	if execErr != nil {
		return fmt.Errorf("error creating template keyspace: %w", execErr)
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
		portInt := cassandraPort
		if port != "" {
			_, _ = fmt.Sscanf(port, "%d", &portInt)
		}

		cluster := cassandra.NewCluster("localhost")
		cluster.Port = portInt
		cluster.ConnectTimeout = pingConnectTimeout
		cluster.Timeout = pingTimeout
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
	port := cassandraPort
	if dbPort != "" {
		_, _ = fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	cluster.ConnectTimeout = dockerConnectTimeout
	cluster.Timeout = dockerTimeout
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
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("httpfs", sourceDriver, connectionString)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
