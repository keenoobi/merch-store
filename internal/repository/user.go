package repository

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	query := `SELECT username, password_hash, coins FROM users WHERE username = $1`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.Name,
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

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3) RETURNING username`
	err := r.db.QueryRow(ctx, query, user.Name, user.PasswordHash, user.Coins).Scan(&user.Name)
	if err != nil {
		slog.Error("Failed to create user", "username", user.Name, "error", err)
		return err
	}

	slog.Info("User successfully created in db", "username", user.Name)
	return nil
}

// TODO: Может сделать под каждое изменение отдельную функцию? Пока сделал для коинов
func (r *UserRepository) UpdateUserCoins(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET coins = $1 WHERE username = $2`
	result, err := r.db.Exec(ctx, query, user.Coins, user.Name)
	if err != nil {
		slog.Error("Failed to update user coins", "username", user.Name, "error", err)
		return err
	}

	// Проверяем, что обновление действительно произошло
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		slog.Error("No rows affected", "username", user.Name)
		return errors.New("no rows affected")
	}

	slog.Info("User coins successfully updated", "username", user.Name)
	return nil
}

// GetUserBalance возвращает баланс пользователя
func (r *UserRepository) GetUserBalance(ctx context.Context, username string) (int, error) {
	var balance int
	query := `SELECT coins FROM users WHERE username = $1`
	err := r.db.QueryRow(ctx, query, username).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("user not found: %s", username)
		}
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}
	return balance, nil
}

// GetUserInventory возвращает инвентарь пользователя
func (r *UserRepository) GetUserInventory(ctx context.Context, username string) ([]entity.InventoryItem, error) {
	query := `SELECT item_name, quantity FROM inventory WHERE user_name = $1`
	rows, err := r.db.Query(ctx, query, username)
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
	return inventory, nil
}
