package usecase

import (
	"avito-merch/internal/entity"
	"avito-merch/internal/repository"
	"context"
	"fmt"
)

type InfoUseCase struct {
	userRepo        *repository.UserRepository
	transactionRepo *repository.TransactionRepository
}

func NewInfoUseCase(userRepo *repository.UserRepository, transactionRepo *repository.TransactionRepository) *InfoUseCase {
	return &InfoUseCase{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (uc *InfoUseCase) GetUserInfo(ctx context.Context, username string) (*entity.InfoData, error) {
	// Получаем баланс пользователя
	balance, err := uc.userRepo.GetUserBalance(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	// Получаем инвентарь пользователя
	inventory, err := uc.userRepo.GetUserInventory(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user inventory: %w", err)
	}

	// Получаем историю транзакций
	transactions, err := uc.transactionRepo.GetTransfersByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	// Формируем ответ
	info := &entity.InfoData{
		Coins:     balance,
		Inventory: inventory,
		CoinHistory: entity.CoinHistory{
			Received: uc.filterReceivedTransactions(transactions, username),
			Sent:     uc.filterSentTransactions(transactions, username),
		},
	}

	return info, nil
}

// filterReceivedTransactions фильтрует полученные транзакции
func (uc *InfoUseCase) filterReceivedTransactions(transactions []entity.Transaction, username string) []entity.Transaction {
	var received []entity.Transaction
	for _, tx := range transactions {
		if tx.ToUser == username {
			received = append(received, entity.Transaction{
				FromUser: tx.FromUser,
				Amount:   tx.Amount,
			})
		}
	}
	return received
}

// filterSentTransactions фильтрует отправленные транзакции
func (uc *InfoUseCase) filterSentTransactions(transactions []entity.Transaction, username string) []entity.Transaction {
	var sent []entity.Transaction
	for _, tx := range transactions {
		if tx.FromUser == username {
			sent = append(sent, entity.Transaction{
				ToUser: tx.ToUser,
				Amount: tx.Amount,
			})
		}
	}
	return sent
}
