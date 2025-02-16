package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/internal/utils"
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
	fromUsername, ok := context.GetUserName(r.Context())
	if !ok {
		slog.Error("User ID not found in context")
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Amount <= 0 {
		slog.Error("Invalid amount", "amount", req.Amount)
		utils.WriteError(w, http.StatusBadRequest, "Amount is required and must be positive")
		return
	}
	if req.ToUser == "" {
		slog.Error("Recipient username is empty")
		utils.WriteError(w, http.StatusBadRequest, "toUser is required")
		return
	}

	// Выполняем перевод
	if err := h.sendCoinUseCase.SendCoins(r.Context(), fromUsername, req.ToUser, req.Amount); err != nil {
		slog.Error("Failed to send coins", "fromUsername", fromUsername, "toUser", req.ToUser, "amount", req.Amount, "error", err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Coins transferred successfully"}); err != nil {
		slog.Error("failed to encode JSON response")
	}
}
