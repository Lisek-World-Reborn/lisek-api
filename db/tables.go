package db

import (
	"strings"
	"time"
)

type User struct {
	ID        uint `gorm:"primary_key"`
	Username  string
	UUID      string
	DiscordID uint64
	CreatedAt time.Time
}

func (u *User) SafeUsername() string {
	return strings.ToLower(strings.Replace(u.Username, " ", "_", -1))
}
