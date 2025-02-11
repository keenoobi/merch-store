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

type ItemRepository struct {
	db *pgxpool.Pool
}

func NewItemRepository(db *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{db: db}
}

// GetItemByName возвращает товар по названию
func (r *ItemRepository) GetItemByName(ctx context.Context, name string) (*entity.Item, error) {
	var item entity.Item
	query := `SELECT name, price FROM merch_items WHERE name = $1`

	err := r.db.QueryRow(ctx, query, name).Scan(
		&item.Name,
		&item.Price,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.Info("Item not found", "name", name)
		return nil, nil
	}
	if err != nil {
		slog.Error("Failed to get item by name", "name", name, "error", err)
		return nil, err
	}

	slog.Info("Item retrieved", "name", name)
	return &item, nil
}

// AddToInventory добавляет товар в инвентарь пользователя
func (r *ItemRepository) AddToInventory(ctx context.Context, userID uuid.UUID, itemName string) error {
	query := `INSERT INTO inventory (user_id, item_name, quantity) 
	VALUES ($1, $2, 1) ON CONFLICT (user_id, item_name) DO UPDATE SET quantity = inventory.quantity + 1`
	_, err := r.db.Exec(ctx, query, userID, itemName)
	if err != nil {
		slog.Error("Failed to add item to inventory", "userID", userID, "item", itemName, "error", err)
		return err
	}

	slog.Info("Item added to inventory", "userID", userID, "item", itemName)
	return nil
}
