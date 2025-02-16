package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/internal/utils"
	"avito-merch/pkg/context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type BuyHandler struct {
	buyUseCase *usecase.BuyUseCase
}

func NewBuyHandler(buyUseCase *usecase.BuyUseCase) *BuyHandler {
	return &BuyHandler{buyUseCase: buyUseCase}
}

func (h *BuyHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
	userName, ok := context.GetUserName(r.Context())
	if !ok {
		slog.Error("User not found in context")
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	itemName := mux.Vars(r)["item"]

	if err := h.buyUseCase.BuyItem(r.Context(), userName, itemName); err != nil {
		slog.Error("Failed to buy item", "userName", userName, "item", itemName, "error", err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "Item purchased successfully"}); err != nil {
		slog.Error("failed to encode JSON response")
	}
}
