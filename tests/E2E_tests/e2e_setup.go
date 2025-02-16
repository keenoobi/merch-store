package e2e

import (
	"avito-merch/internal/handlers"
	"avito-merch/internal/repository"
	"avito-merch/internal/usecase"
	"avito-merch/pkg/auth"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerConfig struct {
	DBName         string
	DBUser         string
	DBPassword     string
	MigrationsPath string
}

type AuthResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Errors string `json:"errors"`
}

type InfoResponse struct {
	Coins       int                 `json:"coins"`
	Inventory   []InventoryItem     `json:"inventory"`
	CoinHistory CoinHistoryResponse `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistoryResponse struct {
	Received []TransactionResponse `json:"received"`
	Sent     []TransactionResponse `json:"sent"`
}

type TransactionResponse struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
	Amount   int    `json:"amount"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type SendCoinResponse struct {
	Message string `json:"message,omitempty"`
	Errors  string `json:"errors,omitempty"`
}

type BuyItemResponse struct {
	Message string `json:"message,omitempty"`
	Errors  string `json:"errors,omitempty"`
}

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	ctx := context.Background()

	migrationsPath, _ := filepath.Abs("../../migrations")

	dsn, cleanup, err := StartPostgresContainer(ctx, &ContainerConfig{
		DBName:         "testdb",
		DBUser:         "user",
		DBPassword:     "password",
		MigrationsPath: migrationsPath,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("Failed to parse database configuration: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = 30 * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("failed to connect to the database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("database is unavailable: %v", err)
	}
	if err != nil {
		t.Fatalf("Failed to connect to the database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authUseCase := usecase.NewAuthUseCase(userRepo)
	buyUseCase := usecase.NewBuyUseCase(userRepo, itemRepo)
	infoUseCase := usecase.NewInfoUseCase(userRepo, transactionRepo)
	sendCoinUseCase := usecase.NewSendCoinUseCase(userRepo, transactionRepo)

	authHandler := handlers.NewAuthHandler(authUseCase)
	buyHandler := handlers.NewBuyHandler(buyUseCase)
	infoHandler := handlers.NewInfoHandler(infoUseCase)
	sendCoinHandler := handlers.NewSendCoinHandler(sendCoinUseCase)

	r := mux.NewRouter()

	authRouter := r.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("", authHandler.Authenticate).Methods(http.MethodPost)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)
	apiRouter.HandleFunc("/buy/{item}", buyHandler.BuyItem).Methods(http.MethodGet)
	apiRouter.HandleFunc("/sendCoin", sendCoinHandler.SendCoins).Methods(http.MethodPost)
	apiRouter.HandleFunc("/info", infoHandler.GetUserInfo).Methods(http.MethodGet)

	server := httptest.NewServer(r)

	return server, func() {
		server.Close()
		cleanup()
	}
}

func StartPostgresContainer(ctx context.Context, cfg *ContainerConfig) (string, func(), error) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:13-alpine",
		postgres.WithDatabase(cfg.DBName),
		postgres.WithUsername(cfg.DBUser),
		postgres.WithPassword(cfg.DBPassword),
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

	if err := runMigrations(dsn, cfg.MigrationsPath); err != nil {
		return "", nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	terminateContainer := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("Failed to terminate PostgreSQL container: %v\n", err)
		}
	}

	return dsn, terminateContainer, nil
}

func runMigrations(dsn, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
