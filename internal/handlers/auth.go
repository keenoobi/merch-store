package handlers

import (
	"avito-merch/internal/usecase"
	"avito-merch/internal/utils"
	"avito-merch/pkg/auth"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

// TODO: Можно добавить отдельный пакет для валидации длины
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid request", "error", err)
		utils.WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Валидирую, что имя пользователя или пароль не пустые
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		slog.Error("Validation failed", "username", req.Username, "password_length", len(req.Password))
		utils.WriteError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	user, err := h.authUseCase.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		slog.Error("Authentication failed", "error", err)
		utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	token, err := auth.GenerateToken(user.Name)
	if err != nil {
		slog.Error("Failed to generate token", "error", err)
		utils.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		slog.Error("failed to encode JSON response")
	}
}
