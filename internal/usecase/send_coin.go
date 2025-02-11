package usecase

import (
	"avito-merch/internal/entity"
	"avito-merch/internal/repository"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type SendCoinUseCase struct {
	userRepo        *repository.UserRepository
	transactionRepo *repository.TransactionRepository
}

func NewSendCoinUseCase(userRepo *repository.UserRepository, transactionRepo *repository.TransactionRepository) *SendCoinUseCase {
	return &SendCoinUseCase{userRepo: userRepo, transactionRepo: transactionRepo}
}

// SendCoins выполняет перевод монет
func (uc *SendCoinUseCase) SendCoins(ctx context.Context, fromUserID uuid.UUID, toUsername string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive: %d", amount)
	}

	tx, err := uc.transactionRepo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	userRepo := repository.UserRepoWithTx(tx)
	transactionRepo := repository.TransactionRepoWithTx(tx)

	// Находим в БД получателя получателя
	toUser, err := userRepo.GetUserByUsername(ctx, toUsername)
	if err != nil {
		return fmt.Errorf("failed to get recipient: %w", err)
	}
	if toUser == nil {
		return fmt.Errorf("recipient not found: %s", toUsername)
	}

	// Получаем отправителя
	fromUser, err := userRepo.GetUserByID(ctx, fromUserID)
	if err != nil {
		return fmt.Errorf("failed to get sender: %w", err)
	}

	// TODO: Не знаю куда это лучше сделать?
	if toUser.Username == fromUser.Username {
		return fmt.Errorf("wrong coins recipient: %s", toUser.Username)
	}

	// Проверяем баланс
	if fromUser.Coins < amount {
		return fmt.Errorf("insufficient coins: userID=%s, amount=%d", fromUserID, amount)
	}

	// Обновляем балансы
	fromUser.Coins -= amount
	toUser.Coins += amount

	if err := userRepo.UpdateUserCoins(ctx, fromUser); err != nil {
		return fmt.Errorf("failed to update sender balance: %w", err)
	}
	if err := userRepo.UpdateUserCoins(ctx, toUser); err != nil {
		return fmt.Errorf("failed to update recipient balance: %w", err)
	}

	// Создаем запись о переводе
	transfer := &entity.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUser.ID,
		Amount:     amount,
	}
	if err := transactionRepo.CreateTransfer(ctx, transfer); err != nil {
		return fmt.Errorf("failed to create transfer record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Coins transferred successfully", "fromUserID", fromUserID, "toUserID", toUser.ID, "amount", amount)
	return nil
}
