package repository

import (
	"avito-merch/internal/entity"
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type ItemRepository struct {
	db DB
}

func NewItemRepository(db DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// WithTx создает новый репозиторий, работающий в рамках транзакции
func ItemRepoWithTx(tx pgx.Tx) *ItemRepository {
	return NewItemRepository(tx)
}

func (r *ItemRepository) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
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
func (r *ItemRepository) AddToInventory(ctx context.Context, userName string, itemName string) error {
	query := `INSERT INTO inventory (user_name, item_name, quantity) 
	VALUES ($1, $2, 1) ON CONFLICT (user_name, item_name) DO UPDATE SET quantity = inventory.quantity + 1`
	_, err := r.db.Exec(ctx, query, userName, itemName)
	if err != nil {
		slog.Error("Failed to add item to inventory", "userName", userName, "item", itemName, "error", err)
		return err
	}

	slog.Info("Item added to inventory", "userName", userName, "item", itemName)
	return nil
}
