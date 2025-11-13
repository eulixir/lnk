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
	_gocqlDefaultKeyspace      = "master"
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
	Folder string
	FS     fs.FS
	Logger migrate.Logger
}

type DockerContainerConfig struct {
	ReuseContainer bool
	Version        string
	Expire         uint
	ContainerName  string
	Migrations     *Migrations
}

type DockerizedGocql struct {
	Port string
}

func StartDockerContainer(cfg DockerContainerConfig) (_ *DockerizedGocql, teardownFn func(), err error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if containerInitialized && concurrentSession != nil {
		return &DockerizedGocql{Port: dbPort}, func() {}, nil
	}

	dockerPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

	// Set longer timeout for Cassandra (it can take 60+ seconds to start)
	dockerPool.MaxWait = 180 * time.Second

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

	// Wait for Cassandra to be ready with retries (Cassandra can take 60+ seconds to fully start)
	// Use a custom retry with exponential backoff instead of dockertest's default
	fmt.Println("Waiting for Cassandra to be ready...")
	startTime := time.Now()
	maxWait := 180 * time.Second
	backoff := 2 * time.Second

	for {
		if time.Since(startTime) > maxWait {
			return nil, nil, fmt.Errorf("cassandra not ready after %v: timeout exceeded", maxWait)
		}

		err = pingCassandraFn(dbPort)()
		if err == nil {
			fmt.Printf("Cassandra is ready after %v\n", time.Since(startTime))
			break
		}

		fmt.Printf("Cassandra not ready yet (attempt after %v): %v\n", time.Since(startTime), err)
		time.Sleep(backoff)
		if backoff < 10*time.Second {
			backoff *= 2 // Exponential backoff up to 10 seconds
		}
	}

	// Give Cassandra additional time to fully initialize after first successful connection
	time.Sleep(3 * time.Second)

	if cfg.Expire != 0 {
		_ = dockerResource.Expire(cfg.Expire)
	}
	// Retry creating session to ensure Cassandra is fully ready
	err = dockerPool.Retry(func() error {
		var retryErr error
		concurrentSession, retryErr = newDB(getCassandraConnString(dbPort, "master"))
		return retryErr
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session after retries: %w", err)
	}

	containerInitialized = true

	if !templateReady {
		if err := setupTemplateDatabase(concurrentSession, cfg.Migrations.FS); err != nil {
			concurrentSession.Close()
			return nil, nil, err
		}
		templateReady = true
	}

	teardownFn = func() {
		if !cfg.ReuseContainer {
			if concurrentSession != nil {
				concurrentSession.Close()
				concurrentSession = nil
			}
			dockerResource.Close()
			containerInitialized = false
			templateReady = false
		}
	}

	return &DockerizedGocql{
		Port: dbPort,
	}, teardownFn, nil
}

// Setup template keyspace and run migrations on it once.
// This template will be used to quickly create test keyspaces.
func setupTemplateDatabase(conn *cassandra.Session, migrationsFs fs.FS) error {
	dbKeyspace := _templateKeyspace

	// Check if template keyspace exists
	var dbCount int
	err := conn.Query("SELECT COUNT(*) FROM system_schema.keyspaces WHERE keyspace_name = ?", dbKeyspace).Scan(&dbCount)
	if err != nil {
		return fmt.Errorf("error checking template keyspace: %w", err)
	}

	// If template already exists, assume it's ready
	if dbCount > 0 {
		return nil
	}

	fmt.Println("Creating template keyspace and running migrations")

	// Create template keyspace
	err = createDB(dbKeyspace, conn)
	if err != nil {
		return fmt.Errorf("error creating template keyspace: %w", err)
	}

	// Wait briefly to ensure keyspace is fully created
	time.Sleep(time.Millisecond * retryIntervalMs)

	// Apply migrations to template
	err = runMigrations(getCassandraConnString(dbPort, dbKeyspace), migrationsFs)
	if err != nil {
		return fmt.Errorf("failed to run migrations on template: %w", err)
	}

	// Wait briefly for migrations to complete
	time.Sleep(time.Millisecond * retryIntervalMs)

	return nil
}

func dropDB(dbKeyspace string, conn *cassandra.Session) error {
	if dbKeyspace == _gocqlDefaultKeyspace || dbKeyspace == _templateKeyspace {
		return nil
	}
	if err := conn.Query(fmt.Sprintf("DROP KEYSPACE IF EXISTS %s", dbKeyspace)).Exec(); err != nil {
		return fmt.Errorf("failed dropping keyspace %s: %w", dbKeyspace, err)
	}
	return nil
}

func createDB(dbKeyspace string, conn *cassandra.Session) error {
	_ = dropDB(dbKeyspace, conn)

	createQuery := fmt.Sprintf(
		"CREATE KEYSPACE %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}",
		dbKeyspace,
	)
	if err := conn.Query(createQuery).Exec(); err != nil {
		return fmt.Errorf("error creating keyspace %s: %w", dbKeyspace, err)
	}

	return nil
}

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
		if container != nil && !container.State.Running { // clean up old stopped container
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

	if !cfg.ReuseContainer {
		return nil, fmt.Errorf(`RunWithOptions failed: %w`, err)
	}

	container, retryErr := tryRecoverExistingContainer(dockerPool, cfg)
	if retryErr != nil {
		return nil, fmt.Errorf(`RunWithOptions.dockerPool.Retry failed: %w`, retryErr)
	}
	if container != nil {
		return &dockertest.Resource{Container: container}, nil
	}

	return nil, fmt.Errorf(`RunWithOptions failed: %w`, err)
}

func tryRecoverExistingContainer(dockerPool *dockertest.Pool, cfg DockerContainerConfig) (*docker.Container, error) {
	var container *docker.Container
	var err error

	err = dockerPool.Retry(func() error {
		container, err = dockerPool.Client.InspectContainer(cfg.ContainerName)
		if container != nil {
			return nil
		}
		return err
	})

	return container, err
}

func pingCassandraFn(port string) func() error {
	return func() error {
		// Create a minimal cluster config for ping
		portInt := 9042
		if port != "" {
			fmt.Sscanf(port, "%d", &portInt)
		}

		cluster := cassandra.NewCluster("localhost")
		cluster.Port = portInt
		cluster.ConnectTimeout = 5 * time.Second // Shorter timeout for faster retries
		cluster.Timeout = 3 * time.Second
		cluster.Consistency = cassandra.One
		// Disable initial host lookup to speed up connection attempts
		cluster.DisableInitialHostLookup = true
		// Reduce retry attempts for faster failure detection
		cluster.NumConns = 1

		conn, err := cluster.CreateSession()
		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		defer conn.Close()
		// Simple query to verify Cassandra is ready
		err = conn.Query("SELECT now() FROM system.local").Exec()
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		return nil
	}
}

func getCassandraConnString(port, keyspace string) string {
	// Format: cassandra://host:port/keyspace?x-multi-statement=true
	// The keyspace must be in the path, not as a query parameter
	return fmt.Sprintf("cassandra://localhost:%s/%s?x-multi-statement=true", port, keyspace)
}

func newDB(_ string) (*cassandra.Session, error) {
	// Use dbPort from package variable (set from docker resource)
	port := 9042
	if dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	// No authentication for test container (standard Cassandra image doesn't enable it by default)
	cluster.ConnectTimeout = 30 * time.Second
	cluster.Timeout = 10 * time.Second
	cluster.Consistency = cassandra.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	return session, nil
}

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
