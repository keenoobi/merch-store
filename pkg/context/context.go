package context

import (
	"context"
)

type contextKey string

const userNameKey contextKey = "userName"

// WithUserName добавляет userName в контекст
func WithUserName(ctx context.Context, userName string) context.Context {
	return context.WithValue(ctx, userNameKey, userName)
}

// GetuserName возвращает userName из контекста
func GetUserName(ctx context.Context) (string, bool) {
	value := ctx.Value(userNameKey)
	if value == nil {
		return "", false
	}
	userName, ok := value.(string)
	return userName, ok
}
