package repository

import (
	"avito-merch/internal/entity"
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TransactionRepository struct {
	db DB
}

func NewTransactionRepository(db DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func TransactionRepoWithTx(tx pgx.Tx) *TransactionRepository {
	return NewTransactionRepository(tx)
}

func (r *TransactionRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

// Создаем запись о переводе
func (r *TransactionRepository) CreateTransfer(ctx context.Context, transfer *entity.Transaction) error {
	query := `INSERT INTO transfer_history (from_user_id, to_user_id, amount)
		VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, transfer.FromUserID, transfer.ToUserID, transfer.Amount).Scan(&transfer.ID)
	if err != nil {
		slog.Error("Failed to create transfer", "error", err)
		return err
	}
	slog.Info("Transfer created", "fromUserID", transfer.FromUserID, "toUserID", transfer.ToUserID, "amount", transfer.Amount)
	return nil
}

// Получаем историю переводов пользователя
func (r *TransactionRepository) GetTransfersByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error) {
	query := `SELECT 
				id, from_user_id,
				to_user_id,
				amount
			FROM transfer_history 
			WHERE from_user_id = $1 OR to_user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		slog.Error("Failed to get transfers", "userID", userID, "error", err)
		return nil, err
	}
	defer rows.Close()

	var transfers []entity.Transaction
	for rows.Next() {
		var transfer entity.Transaction
		if err := rows.Scan(&transfer.ID, &transfer.FromUserID, &transfer.ToUserID, &transfer.Amount); err != nil {
			slog.Error("Failed to scan transfer", "userID", userID, "error", err)
			return nil, err
		}
		transfers = append(transfers, transfer)
	}

	slog.Info("Transfers retrieved", "userID", userID)
	return transfers, nil
}

// internal/repository/transaction.go
func (r *TransactionRepository) GetUserTransactionHistory(ctx context.Context, userID uuid.UUID) ([]entity.Transaction, error) {
	query := `
        SELECT 
            from_user_id, 
            to_user_id, 
            amount 
        FROM transfer_history 
        WHERE from_user_id = $1 OR to_user_id = $1
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	defer rows.Close()

	var transactions []entity.Transaction
	for rows.Next() {
		var fromUserID, toUserID uuid.UUID
		var amount int
		if err := rows.Scan(&fromUserID, &toUserID, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		transaction := entity.Transaction{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			Amount:     amount,
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
