package usecases_test

import (
	"log"
	"os"
	"testing"

	gocqltesting "lnk/extensions/gocqltesting"
	"lnk/gateways/gocql/migrations"
)

func TestMain(m *testing.M) {
	tearDown, err := gocqltesting.StartDockerContainer(gocqltesting.DockerContainerConfig{
		Version:        "latest",
		ContainerName:  "lnk-cassandra-test",
		ReuseContainer: true,
		Migrations: &gocqltesting.Migrations{
			FS: migrations.MigrationsFS,
		},
	})
	if err != nil {
		log.Fatalf("Failed to start Docker container: %v", err)
	}

	exitVal := m.Run()

	tearDown()

	os.Exit(exitVal)
}
