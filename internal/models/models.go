package models

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password_hash"`
	Coins    int    `json:"coins"`
}

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Items struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type ReceivedCoins struct {
	FromUser string `json:"f"`
	Amount   int    `json:"amount"`
}

type SentCoins struct {
	ToUser string `json:"to_user"`
	Amount int    `json:"amount"`
}
