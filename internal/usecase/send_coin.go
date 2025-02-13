package usecase

import (
	"avito-merch/internal/entity"
	"avito-merch/internal/repository"
	"context"
	"fmt"
	"log/slog"
)

type SendCoinUseCase struct {
	userRepo        *repository.UserRepository
	transactionRepo *repository.TransactionRepository
}

func NewSendCoinUseCase(userRepo *repository.UserRepository, transactionRepo *repository.TransactionRepository) *SendCoinUseCase {
	return &SendCoinUseCase{userRepo: userRepo, transactionRepo: transactionRepo}
}

// SendCoins выполняет перевод монет
func (uc *SendCoinUseCase) SendCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	if fromUsername == toUsername {
		return fmt.Errorf("cannot send coins to yourself: %s", toUsername)
	}

	tx, err := uc.transactionRepo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userRepo := repository.UserRepoWithTx(tx)
	transactionRepo := repository.TransactionRepoWithTx(tx)

	// Обновляем балансы обоих пользователей
	if err := userRepo.UpdateUserAfterTransfer(ctx, fromUsername, toUsername, amount); err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}

	// Создаем запись о переводе
	if err := transactionRepo.CreateTransfer(ctx, &entity.Transaction{
		FromUser: fromUsername,
		ToUser:   toUsername,
		Amount:   amount,
	}); err != nil {
		return fmt.Errorf("failed to create transfer record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Логируем успешный перевод
	slog.Info("Coins transferred successfully",
		"fromUserName", fromUsername,
		"toUserName", toUsername,
		"amount", amount,
	)

	return nil
}
