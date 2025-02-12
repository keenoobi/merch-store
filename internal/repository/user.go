package repository

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db DB
}

func NewUserRepository(db DB) *UserRepository {
	return &UserRepository{db: db}
}

// WithTx создает новый репозиторий, работающий в рамках транзакции
func UserRepoWithTx(tx pgx.Tx) *UserRepository {
	return NewUserRepository(tx)
}

func (r *UserRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, username, password_hash, coins FROM users WHERE username = $1`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Coins,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.Info("User not found", "username", username)
		return nil, nil
	}
	if err != nil {
		slog.Error("Failed to get user by username", "username", username, "error", err)
		return nil, err
	}

	slog.Info("User retrieved", "username", username)
	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, username, password_hash, coins FROM users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Coins,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.Info("User not found", "userID", userID)
		return nil, nil
	}
	if err != nil {
		slog.Error("Failed to get user by userID", "userID", userID, "error", err)
		return nil, err
	}

	slog.Info("User retrieved", "userID", userID)
	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Coins).Scan(&user.ID)
	if err != nil {
		slog.Error("Failed to create user", "username", user.Username, "error", err)
		return err
	}

	slog.Info("User successfully created in db", "username", user.Username)
	return nil
}

// TODO: Может сделать под каждое изменение отдельную функцию? Пока сделал для коинов
func (r *UserRepository) UpdateUserCoins(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET coins = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, user.Coins, user.ID)
	if err != nil {
		slog.Error("Failed to update user coins", "username", user.Username, "error", err)
		return err
	}

	// Проверяем, что обновление действительно произошло
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Error("No rows affected", "username", user.Username)
		return errors.New("no rows affected")
	}

	slog.Info("User coins successfully updated", "username", user.Username)
	return nil
}

// internal/repository/user.go
func (r *UserRepository) GetUserInfo(ctx context.Context, userID uuid.UUID) (*entity.InfoData, error) {
	// Получаем баланс
	var coins int
	err := r.db.QueryRow(ctx, "SELECT coins FROM users WHERE id = $1", userID).Scan(&coins)
	if err != nil {
		return nil, fmt.Errorf("failed to get user coins: %w", err)
	}

	// Получаем инвентарь
	rows, err := r.db.Query(ctx, "SELECT item_name AS type, quantity FROM inventory WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user inventory: %w", err)
	}
	defer rows.Close()

	var inventory []entity.InventoryItem
	for rows.Next() {
		var item entity.InventoryItem
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}
		inventory = append(inventory, item)
	}

	// Получаем полученные транзакции
	rows, err = r.db.Query(ctx, `
        SELECT fu.username AS from_user, th.amount 
        FROM transfer_history th
        JOIN users fu ON th.from_user_id = fu.id
        WHERE th.to_user_id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get received transactions: %w", err)
	}
	defer rows.Close()

	var received []entity.InfoTransaction
	for rows.Next() {
		var transaction entity.InfoTransaction
		if err := rows.Scan(&transaction.FromUser, &transaction.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan received transaction: %w", err)
		}
		received = append(received, transaction)
	}

	// Получаем отправленные транзакции
	rows, err = r.db.Query(ctx, `
        SELECT tu.username AS to_user, th.amount 
        FROM transfer_history th
        JOIN users tu ON th.to_user_id = tu.id
        WHERE th.from_user_id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent transactions: %w", err)
	}
	defer rows.Close()

	var sent []entity.InfoTransaction
	for rows.Next() {
		var transaction entity.InfoTransaction
		if err := rows.Scan(&transaction.ToUser, &transaction.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan sent transaction: %w", err)
		}
		sent = append(sent, transaction)
	}

	return &entity.InfoData{
		Coins:     coins,
		Inventory: inventory,
		CoinHistory: entity.CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}, nil
}
