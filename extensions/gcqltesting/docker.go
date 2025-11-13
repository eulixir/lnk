package gcqltesting

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"sync"

	cassandra "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cassandra"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	migrationHTTPS "github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

const (
	_reusableContainerName     = "gcql-testing-container"
	_defaultGocqlDockerVersion = "latest"
	_templateKeyspace          = "template_keyspace"
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

	FS fs.FS

	Logger migrate.Logger
}

type DockerContainerConfig struct {
	ReuseContainer bool

	Version string

	Expire uint

	ContainerName string

	Migrations *Migrations
}

type DockerizedMssql struct {
	Port string
}

func StartDockerContainer(cfg DockerContainerConfig) (_ *DockerizedMssql, teardownFn func(), err error) {
	setupMutex.Lock()
	defer setupMutex.Unlock()

	if containerInitialized && concurrentSession != nil {
		return &DockerizedMssql{Port: dbPort}, func() {}, nil
	}

	dockerPool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf(`could not connect to docker: %w`, err)
	}

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

	if err = dockerPool.Retry(pingCassandraFn(dbPort)); err != nil {
		return nil, nil, err
	}

	if cfg.Expire != 0 {
		_ = dockerResource.Expire(cfg.Expire)
	}
	concurrentSession, err = newDB(getCassandraConnString(dbPort, "master"))
	if err != nil {
		return nil, nil, err
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

	return &DockerizedMssql{
		Port: dbPort,
	}, teardownFn, nil
}

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
	err = createDB(dbKeyspace, conn)
	if err != nil {
		return fmt.Errorf("error creating template keyspace: %w", err)
	}

	err = runMigrations(getCassandraConnString(dbPort, dbKeyspace), migrationsFs)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func dropDB(dbKeyspace string, conn *cassandra.Session) error {
	if dbKeyspace == _templateKeyspace {
		return nil
	}
	if err := conn.Query(fmt.Sprintf(`DROP KEYSPACE IF EXISTS [%s]`, dbKeyspace)).Exec(); err != nil {
		return fmt.Errorf("failed dropping keyspace %s: %w", dbKeyspace, err)
	}
	return nil
}

func createDB(dbKeyspace string, conn *cassandra.Session) error {
	_ = dropDB(dbKeyspace, conn)

	if err := conn.Query(fmt.Sprintf(`CREATE KEYSPACE [%s]`, dbKeyspace)).Exec(); err != nil {
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
		Repository: "cassandra/cassandra",
		Tag:        cfg.Version,
		Env:        []string{"CASSANDRA_CLUSTER_NAME=lnk-cluster", "CASSANDRA_DC=datacenter1", "CASSANDRA_RACK=rack1", "CASSANDRA_AUTHENTICATOR=PasswordAuthenticator", "CASSANDRA_AUTHORIZER=CassandraAuthorizer", "CASSANDRA_USER=cassandra", "CASSANDRA_PASSWORD=cassandra"},
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
		conn, err := newDB(getCassandraConnString(port, "master"))
		if err != nil {
			return err
		}

		defer conn.Close()
		// Test connection by querying system keyspaces
		err = conn.Query("SELECT keyspace_name FROM system_schema.keyspaces LIMIT 1").Exec()
		return err
	}
}

func getCassandraConnString(port, dbName string) string {
	query := url.Values{}
	query.Add("TrustServerCertificate", "true")

	return fmt.Sprintf("cassandra://localhost:%s?database=%s", port, dbName)
}

func newDB(connectionURL string) (*cassandra.Session, error) {
	// Use dbPort from package variable (set from docker resource)
	port := 9042
	if dbPort != "" {
		fmt.Sscanf(dbPort, "%d", &port)
	}

	cluster := cassandra.NewCluster("localhost")
	cluster.Port = port
	cluster.Authenticator = cassandra.PasswordAuthenticator{
		Username: "cassandra",
		Password: "cassandra",
	}
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
