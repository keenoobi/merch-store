package usecase

import (
	"avito-merch/internal/repository"
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type BuyUseCase struct {
	userRepo *repository.UserRepository
	itemRepo *repository.ItemRepository
}

func NewBuyUseCase(userRepo *repository.UserRepository, itemRepo *repository.ItemRepository) *BuyUseCase {
	return &BuyUseCase{userRepo: userRepo, itemRepo: itemRepo}
}

// BuyItem выполняет покупку товара
func (uc *BuyUseCase) BuyItem(userID uuid.UUID, itemName string) error {
	// Получаем товар
	item, err := uc.itemRepo.GetItemByName(context.Background(), itemName)
	if err != nil {
		slog.Error("Failed to get item", "item", itemName, "error", err)
		return err
	}
	if item == nil {
		slog.Error("Item not found", "item", itemName)
		return errors.New("item not found")
	}

	// Получаем пользователя
	user, err := uc.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		slog.Error("Failed to get user", "userID", userID, "error", err)
		return err
	}

	// Проверяем баланс
	if user.Coins < item.Price {
		slog.Error("Insufficient coins", "userID", userID, "item", itemName)
		return errors.New("insufficient coins")
	}

	// Обновляем баланс пользователя
	user.Coins -= item.Price
	if err := uc.userRepo.UpdateUserCoins(context.Background(), user); err != nil {
		slog.Error("Failed to update user balance", "userID", userID, "error", err)
		return err
	}

	// Добавляем товар в инвентарь
	if err := uc.itemRepo.AddToInventory(context.Background(), userID, itemName); err != nil {
		slog.Error("Failed to add item to inventory", "userID", userID, "item", itemName, "error", err)
		return err
	}

	slog.Info("Item purchased successfully", "userID", userID, "item", itemName)
	return nil
}
