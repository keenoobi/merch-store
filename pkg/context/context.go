package context

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

// WithUserID добавляет userID в контекст
func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID возвращает userID из контекста
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}
