package testutils

import (
	"context"
	"fmt"
	"github.com/basemind-ai/monorepo/shared/go/db"
	"github.com/rs/zerolog/log"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func CreateNamespaceTestDBModule(namespace string) func() {
	connectionPool, poolInitErr := dockertest.NewPool("")
	if poolInitErr != nil {
		log.Fatal().Err(poolInitErr).Msg("failed to construct dockertest pool")
	}

	if poolErr := connectionPool.Client.Ping(); poolErr != nil {
		log.Fatal().Err(poolErr).Msg("failed to connect to docker daemon")
	}

	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=test",
			fmt.Sprintf("POSTGRES_DB=%s", namespace),
			"listen_addresses = '*'",
		},
	}

	resource, dockerErr := connectionPool.RunWithOptions(opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if dockerErr != nil {
		log.Fatal().Err(dockerErr).Msg("failed to start docker container")
	}

	_ = resource.Expire(120)

	dbURL := fmt.Sprintf(
		"postgres://test:test@%s/%s?sslmode=disable",
		resource.GetHostPort("5432/tcp"),
		namespace,
	)

	connection, connectionErr := db.CreateConnection(context.TODO(), dbURL)
	if connectionErr != nil {
		log.Fatal().Err(connectionErr).Msg("failed to connect to test database")
	}

	if migrationsErr := migrateTestDB(dbURL); migrationsErr != nil {
		_ = connectionPool.Purge(resource)
		log.Fatal().Err(migrationsErr).Msg("failed to migrate test database")
	}

	return func() {
		connection.Close()
		if purgeErr := connectionPool.Purge(resource); purgeErr != nil {
			log.Fatal().Err(purgeErr).Msg("failed to purge test database")
		}
	}
}

func migrateTestDB(dbURL string) error {
	_, filePath, _, _ := runtime.Caller(1)
	repositoryRoot, _ := filepath.Abs(filePath + "/../../../../")
	migrationsPath := fmt.Sprintf("file://%s/sql/migrations", repositoryRoot)

	cmd := exec.Command(
		"atlas",
		"migrate",
		"apply",
		"--url",
		dbURL,
		"--dir",
		migrationsPath,
	)

	_, migrationCommandErr := cmd.Output()
	return migrationCommandErr
}
