package auth

import (
	"avito-merch/internal/utils"
	"avito-merch/pkg/context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenString)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Используем кастомный тип для ключа контекста
		ctx := context.WithUserName(r.Context(), claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
