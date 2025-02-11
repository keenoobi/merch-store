package repository

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
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
