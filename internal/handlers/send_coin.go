package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/pkg/context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type SendCoinHandler struct {
	sendCoinUseCase *usecase.SendCoinUseCase
}

func NewSendCoinHandler(sendCoinUseCase *usecase.SendCoinUseCase) *SendCoinHandler {
	return &SendCoinHandler{sendCoinUseCase: sendCoinUseCase}
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func (h *SendCoinHandler) SendCoins(w http.ResponseWriter, r *http.Request) {
	var req SendCoinRequest
	// Получаем userID из контекста от middleware
	fromUserID, ok := context.GetUserID(r.Context())
	if !ok {
		slog.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request", "error", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Валидация TODO: Где её лучше сделать, здесь или в usecase?
	if req.Amount <= 0 {
		slog.Error("Invalid amount", "amount", req.Amount)
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}
	if req.ToUser == "" {
		slog.Error("Recipient username is empty")
		http.Error(w, "Recipient username is required", http.StatusBadRequest)
		return
	}

	// Выполняем перевод
	if err := h.sendCoinUseCase.SendCoins(r.Context(), fromUserID, req.ToUser, req.Amount); err != nil {
		slog.Error("Failed to send coins", "fromUserID", fromUserID, "toUser", req.ToUser, "amount", req.Amount, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Coins transferred successfully"))
}
