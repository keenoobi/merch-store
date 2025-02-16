package app

import (
	"avito-merch/internal/handlers"
	"avito-merch/pkg/auth"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func setupRouter(
	authHandler *handlers.AuthHandler,
	buyHandler *handlers.BuyHandler,
	sendCoinHandler *handlers.SendCoinHandler,
	infoHandler *handlers.InfoHandler,
) *mux.Router {
	r := mux.NewRouter()

	// Регистрируем эндпоинт для аутентификации
	authRouter := r.PathPrefix("/api/auth").Subrouter()
	authRouter.HandleFunc("", authHandler.Authenticate).Methods(http.MethodPost)

	// Регистрируем защищенные эндпоинты
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.AuthMiddleware)
	apiRouter.HandleFunc("/buy/{item}", buyHandler.BuyItem).Methods(http.MethodGet)
	apiRouter.HandleFunc("/sendCoin", sendCoinHandler.SendCoins).Methods(http.MethodPost)
	apiRouter.HandleFunc("/info", infoHandler.GetUserInfo).Methods(http.MethodGet)

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			slog.Error("failed to write response")
		}
	}).Methods("GET")

	return r
}
