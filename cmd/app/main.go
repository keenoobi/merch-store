package main

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
	"avito-merch/pkg/auth"
	"avito-merch/pkg/database"

	"github.com/gorilla/mux"
)

func main() {
	// Инициализация логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Подключаемся к БД
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("Failed to connect to the database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Инициализируем репозитории
	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	// Инициализируем usecases
	authUseCase := usecase.NewAuthUseCase(userRepo)
	buyUseCase := usecase.NewBuyUseCase(userRepo, itemRepo)
	sendCoinUseCase := usecase.NewSendCoinUseCase(userRepo, transactionRepo)

	// Инициализируем handlers
	authHandler := handlers.NewAuthHandler(authUseCase)
	buyHandler := handlers.NewBuyHandler(buyUseCase)
	sendCoinHandler := handlers.NewSendCoinHandler(sendCoinUseCase)

	// Настраиваем роутер
	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Регистрируем эндпоинт для аутентификации
	authRouter := r.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("", authHandler.Authenticate).Methods(http.MethodPost)

	// Регистрируем защищенные эндпоинты
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)
	apiRouter.HandleFunc("/buy/{item}", buyHandler.BuyItem).Methods(http.MethodGet)
	apiRouter.HandleFunc("/send-coin", sendCoinHandler.SendCoins).Methods(http.MethodPost)

	// Запускаем сервер
	serverPort := cfg.ServerPort
	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Горутина для graceful shutdown
	go func() {
		slog.Info("Server started", "port", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидаем сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down the service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown the server", "error", err)
		os.Exit(1)
	}
	slog.Info("Service stopped gracefully")
}
