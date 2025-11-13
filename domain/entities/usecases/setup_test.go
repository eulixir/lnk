package usecases_test

import (
	gocqltesting "lnk/extensions/gocqltesting"
	"lnk/gateways/gocql/migrations"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_, tearDown, err := gocqltesting.StartDockerContainer(gocqltesting.DockerContainerConfig{
		Version:       "latest",
		ContainerName: "lnk-cassandra-test",
		Migrations: &gocqltesting.Migrations{
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
