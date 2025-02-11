package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
}

type AuthUseCase struct {
	userRepo UserRepository
}

func NewAuthUseCase(userRepo UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo}
}

func (uc *AuthUseCase) Authenticate(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		slog.Error("Failed to get user by username", "username", username, "error", err)
		return nil, err
	}
	if user == nil {
		// Создаем нового пользователя TODO: Вынести в отдельную функцию типа registerNewUser
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("Failed to hash password", "error", err)
			return nil, err
		}
		user = &entity.User{
			Username:     username,
			PasswordHash: string(hashedPassword),
			Coins:        1000, // TODO: Вынести в кофиг, или просто const?
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			slog.Error("Failed to create user", "username", username, "error", err)
			return nil, err
		}
		slog.Info("New user created", "username", username)
	} else {
		// Проверяем пароль
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			slog.Error("Invalid password", "username", username, "error", err)
			return nil, err
		}
	}
	return user, nil
}
