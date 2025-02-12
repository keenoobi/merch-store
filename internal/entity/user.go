package entity

type User struct {
	Name         string `json:"username"`
	PasswordHash string `json:"-"`
	Coins        int    `json:"coins"` // TODO: Поменять на balance?
}
