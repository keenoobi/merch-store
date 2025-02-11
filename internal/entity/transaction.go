package entity

import "github.com/google/uuid"

type Transaction struct {
	ID         uuid.UUID `json:"id"`
	FromUserID uuid.UUID `json:"from_user_id"`
	ToUserID   uuid.UUID `json:"to_user_id"`
	Amount     int       `json:"amount"`
}
