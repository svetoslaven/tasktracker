package models

import "time"

type User struct {
	ID           int64  `json:"-"`
	Username     string `json:"name"`
	Email        string `json:"email"`
	PasswordHash []byte `json:"-"`
	IsVerified   bool   `json:"is_verified"`
	Version      int    `json:"-"`
}

type Token struct {
	Plaintext   string     `json:"token"`
	Hash        []byte     `json:"-"`
	RecipientID int64      `json:"-"`
	ExpiresAt   time.Time  `json:"expires_at"`
	Scope       TokenScope `json:"-"`
}

type Team struct {
	ID       int64  `json:"-"`
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
	Version  int    `json:"-"`
}
