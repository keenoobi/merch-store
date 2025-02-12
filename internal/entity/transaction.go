package entity

import "github.com/google/uuid"

type Transaction struct {
	ID         uuid.UUID `json:"id"`
	FromUserID uuid.UUID `json:"from_user_id"`
	ToUserID   uuid.UUID `json:"to_user_id"`
	Amount     int       `json:"amount"`
}

type InfoTransaction struct {
	FromUser string `json:"fromUser,omitempty"` // omitempty, если поле пустое
	ToUser   string `json:"toUser,omitempty"`   // omitempty, если поле пустое
	Amount   int    `json:"amount"`
}
