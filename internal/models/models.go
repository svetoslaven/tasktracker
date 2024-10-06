package models

type User struct {
	ID           int64  `json:"-"`
	Username     string `json:"name"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	IsVerified   bool   `json:"is_verified"`
	Version      int    `json:"-"`
}
