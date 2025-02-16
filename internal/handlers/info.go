package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/internal/utils"
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
	w.Header().Set("Content-Type", "application/json")

	userName, ok := context.GetUserName(r.Context())
	if !ok {
		slog.Error("User not found in context")
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	info, err := h.infoUseCase.GetUserInfo(r.Context(), userName)
	if err != nil {
		slog.Error("Failed to get user info", "userName", userName, "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Failed to get user info")
		return
	}

	if err := json.NewEncoder(w).Encode(info); err != nil {
		slog.Error("Failed to encode response", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
	}
	w.WriteHeader(http.StatusOK)
}
