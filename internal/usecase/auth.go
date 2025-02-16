package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"log/slog"
)

const coins = 1000

type UserRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	GetUserInventory(ctx context.Context, username string) ([]entity.InventoryItem, error)
	UpdateUserAfterPurchase(ctx context.Context, username string, amount int) error
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
		user = &entity.User{
			Name:     username,
			Password: password,
			Coins:    coins,
		}
		if err := uc.userRepo.Create(ctx, user); err != nil {
			slog.Error("Failed to create user", "username", username, "error", err)
			return nil, err
		}
		slog.Info("New user created", "username", username)
	}

	return user, nil
}
