package testutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DBName         = "testdb"
	DBUser         = "user"
	DBPassword     = "password"
	MigrationsPath = "migrations"
)

// StartPostgresContainer запускает контейнер с PostgreSQL и применяет миграции.
func StartPostgresContainer(ctx context.Context) (string, func(), error) {
	// Запускаем контейнер с PostgreSQL
	postgresContainer, err := postgres.Run(ctx,
		"postgres:13-alpine",
		postgres.WithDatabase(DBName),
		postgres.WithUsername(DBUser),
		postgres.WithPassword(DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", MigrationsPath),
		dsn,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return "", nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	terminateContainer := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate PostgreSQL container: %v\n", err)
		}
	}

	// Возвращаем строку подключения и функцию для остановки контейнера
	return dsn, terminateContainer, nil
}
