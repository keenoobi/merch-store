package usecase

import (
	"avito-merch/internal/entity"
	"context"
	"fmt"
)

type TransactionRepository interface {
	GetTransfersByUsername(ctx context.Context, username string) ([]entity.Transaction, error)
}

type InfoUseCase struct {
	userRepo        UserRepository
	transactionRepo TransactionRepository
}

func NewInfoUseCase(userRepo UserRepository, transactionRepo TransactionRepository) *InfoUseCase {
	return &InfoUseCase{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
	}
}

func (uc *InfoUseCase) GetUserInfo(ctx context.Context, username string) (*entity.InfoData, error) {
	// Получаем баланс пользователя
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
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
		Coins:     user.Coins,
		Inventory: inventory,
		CoinHistory: entity.CoinHistory{
			Received: uc.filterReceivedTransactions(transactions, username),
			Sent:     uc.filterSentTransactions(transactions, username),
		},
	}

	if info.Inventory == nil {
		info.Inventory = []entity.InventoryItem{}
	}

	if info.CoinHistory.Received == nil {
		info.CoinHistory.Received = []entity.Transaction{}
	}
	if info.CoinHistory.Sent == nil {
		info.CoinHistory.Sent = []entity.Transaction{}
	}

	return info, nil
}

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
