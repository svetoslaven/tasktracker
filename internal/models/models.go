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

type Invitation struct {
	ID      int64 `json:"id"`
	Team    *Team `json:"team"`
	Inviter *User `json:"inviter"`
	Invitee *User `json:"invitee"`
}

type Membership struct {
	TeamID     int64      `json:"-"`
	Member     *User      `json:"member"`
	MemberRole MemberRole `json:"role"`
	Version    int        `json:"-"`
}

type Task struct {
	ID          int64        `json:"id"`
	CreatedAt   time.Time    `json:"created_at"`
	Due         time.Time    `json:"due"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Creator     *User        `json:"creator"`
	Assignee    *User        `json:"assignee"`
	Version     int          `json:"-"`
}
