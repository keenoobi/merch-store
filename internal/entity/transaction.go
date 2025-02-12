package entity

import "github.com/google/uuid"

type Transaction struct {
	ID       uuid.UUID `json:"-"`
	FromUser string    `json:"FromUser,omitempty"`
	ToUser   string    `json:"ToUser,omitempty"`
	Amount   int       `json:"amount"`
}
