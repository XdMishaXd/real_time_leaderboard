package models

import "time"

type User struct {
	ID         int64
	Email      string
	Username   string
	PassHash   []byte
	IsVerified bool
}

type App struct {
	ID     int32
	Name   string
	Secret string
}

type RefreshToken struct {
	TokenHash []byte
	UserID    int64
	AppID     int32
	ExpiresAt time.Time
}

type Message struct {
	Email string `json:"to"`
	Link  string `json:"link"`
}
