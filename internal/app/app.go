package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avito-merch/internal/config"
	"avito-merch/internal/handlers"
	"avito-merch/internal/repository"
	"avito-merch/internal/usecase"
	"avito-merch/pkg/database"

	"github.com/gorilla/mux"
)

type App struct {
	cfg    *config.Config
	router *mux.Router
	server *http.Server
}

func NewApp(cfg *config.Config) *App {
	// Инициализируем логгер
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	slog.SetDefault(logger)

	// Подключаемся к БД
	db, err := database.NewPostgresDB(cfg.DBConfig)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Инициализируем usecases
	authUseCase := usecase.NewAuthUseCase(userRepo)
	buyUseCase := usecase.NewBuyUseCase(userRepo, itemRepo)
	sendCoinUseCase := usecase.NewSendCoinUseCase(userRepo, transactionRepo)
	infoUseCase := usecase.NewInfoUseCase(userRepo, transactionRepo)

	// Инициализируем handlers
	authHandler := handlers.NewAuthHandler(authUseCase)
	buyHandler := handlers.NewBuyHandler(buyUseCase)
	sendCoinHandler := handlers.NewSendCoinHandler(sendCoinUseCase)
	infoHandler := handlers.NewInfoHandler(infoUseCase)

	// Настраиваем роутер
	router := setupRouter(authHandler, buyHandler, sendCoinHandler, infoHandler)

	// Инициализируем сервер
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return &App{
		cfg:    cfg,
		router: router,
		server: server,
	}
}

func (a *App) Run() {
	// Горутина для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("Server started", "port", a.cfg.ServerPort)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидаем сигнала завершения
	<-stop

	slog.Info("Shutting down the service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown the server", "error", err)
		os.Exit(1)
	}
	slog.Info("Service stopped gracefully")
}
