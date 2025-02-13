package usecase

import (
	"avito-merch/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type BuyUseCase struct {
	userRepo *repository.UserRepository
	itemRepo *repository.ItemRepository
}

func NewBuyUseCase(userRepo *repository.UserRepository, itemRepo *repository.ItemRepository) *BuyUseCase {
	return &BuyUseCase{userRepo: userRepo, itemRepo: itemRepo}
}

// BuyItem выполняет покупку товара
func (uc *BuyUseCase) BuyItem(ctx context.Context, userName string, itemName string) error {
	// Начинаем транзакцию
	tx, err := uc.itemRepo.Begin(ctx)
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userRepo := repository.UserRepoWithTx(tx)
	itemRepo := repository.ItemRepoWithTx(tx)

	// Получаем цену товара
	item, err := itemRepo.GetItemByName(ctx, itemName)
	if err != nil {
		slog.Error("Failed to get item", "item", itemName, "error", err)
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		slog.Error("Item not found", "item", itemName)
		return fmt.Errorf("item not found: %s", itemName)
	}

	// Обновляем баланс пользователя с проверкой

	if err := userRepo.UpdateUserAfterPurchase(ctx, userName, item.Price); err != nil {
		slog.Error("Failed to update user balance", "userName", userName, "error", err)
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	// Добавляем товар в инвентарь
	if err := itemRepo.AddToInventory(ctx, userName, itemName); err != nil {
		slog.Error("Failed to add item to inventory", "userName", userName, "item", itemName, "error", err)
		return fmt.Errorf("failed to add item to inventory: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Item purchased successfully", "userName", userName, "item", itemName)
	return nil
}
