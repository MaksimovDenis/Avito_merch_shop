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
