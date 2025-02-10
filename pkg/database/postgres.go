package database

import (
	"avito-merch/internal/config"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresDB создаёт подключение к БД через pgx
func NewPostgresDB(cfg *config.Config) (*pgxpool.Pool, error) {
	dbHost := cfg.DBHost
	dbPort := cfg.DBPort
	dbUser := cfg.DBUser
	dbPass := cfg.DBPassword
	dbName := cfg.DBName

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		slog.Error("Failed to parse database configuration", "error", err)
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}

	// Устанавливаем параметры пула соединений
	config.MaxConns = 10                        // Максимум 10 подключений
	config.MinConns = 2                         // Минимум 2 подключения
	config.MaxConnLifetime = time.Hour          // Закрытие соединения через 1 час
	config.HealthCheckPeriod = 30 * time.Second // Проверка соединений каждые 30 сек

	// Создаём пул соединений
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		db.Close()
		slog.Error("Database is unavailable", "error", err)
		return nil, fmt.Errorf("database is unavailable: %w", err)
	}

	slog.Info("Successfully connected to PostgreSQL (pgx)")
	return db, nil
}
