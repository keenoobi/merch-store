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
func (uc *BuyUseCase) BuyItem(ctx context.Context, userID uuid.UUID, itemName string) error {
	tx, err := uc.itemRepo.Begin(ctx)
	if err != nil {
		slog.Error("Failed to begin transactions", "error", err)
		return err
	}
	defer tx.Rollback(ctx)

	userRepo := repository.UserRepoWithTx(tx)
	itemRepo := repository.ItemRepoWithTx(tx)

	// Получаем товар
	item, err := itemRepo.GetItemByName(ctx, itemName)
	if err != nil {
		slog.Error("Failed to get item", "item", itemName, "error", err)
		return err
	}
	if item == nil {
		slog.Error("Item not found", "item", itemName)
		return errors.New("item not found")
	}

	// Получаем пользователя
	user, err := userRepo.GetUserByID(ctx, userID)
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
	if err := userRepo.UpdateUserCoins(ctx, user); err != nil {
		slog.Error("Failed to update user balance", "userID", userID, "error", err)
		return err
	}

	// Добавляем товар в инвентарь
	if err := itemRepo.AddToInventory(ctx, userID, itemName); err != nil {
		slog.Error("Failed to add item to inventory", "userID", userID, "item", itemName, "error", err)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return err
	}

	slog.Info("Item purchased successfully", "userID", userID, "item", itemName)
	return nil
}
