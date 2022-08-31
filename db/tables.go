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

type Server struct {
	ID            uint      `gorm:"primary_key" json:"id"`
	Name          string    `json:"name"`
	IP            string    `json:"ip"`
	Port          int       `json:"port"`
	Region        string    `json:"region"`
	CreatedAt     time.Time `json:"created_at"`
	LastPing      time.Time `json:"last_ping"`
	ContainerName string    `json:"container_name"`
}
