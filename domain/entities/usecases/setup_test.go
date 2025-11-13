package usecases_test

import (
	"lnk/extensions/gcqltesting"
	"lnk/gateways/gocql/migrations"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_, tearDown, err := gcqltesting.StartDockerContainer(gcqltesting.DockerContainerConfig{
		Version:       "latest",
		ContainerName: "lnk-cassandra-test",
		Migrations: &gcqltesting.Migrations{
			Folder: "migrations",
			FS:     migrations.MigrationsFS,
		},
	})
	if err != nil {
		log.Fatalf("Failed to start Docker container: %v", err)
	}

	exitVal := m.Run()

	tearDown()

	os.Exit(exitVal)
}
