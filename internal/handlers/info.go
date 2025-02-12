// internal/handlers/info.go
package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/pkg/context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type InfoHandler struct {
	infoUseCase *usecase.InfoUseCase
}

func NewInfoHandler(infoUseCase *usecase.InfoUseCase) *InfoHandler {
	return &InfoHandler{infoUseCase: infoUseCase}
}

func (h *InfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, ok := context.GetUserID(r.Context())
	if !ok {
		slog.Error("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем информацию о пользователе
	info, err := h.infoUseCase.GetUserInfo(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to get user info", "userID", userID, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
