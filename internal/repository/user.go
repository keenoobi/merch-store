package repository

import (
	"avito-merch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"log/slog"

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

	if errors.Is(err, sql.ErrNoRows) {
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
	query := `INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, user.Username, user.PasswordHash, user.Coins).Scan(&user.ID)
	if err != nil {
		slog.Error("Failed to create user", "username", user.Username, "error", err)
		return err
	}

	slog.Info("User created", "username", user.Username)
	return nil
}
