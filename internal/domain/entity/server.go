package entity

import (
	"time"

	"gorm.io/gorm"
)

type ServerStatus string

const (
	ServerStatusUnknown ServerStatus = "UNKNOWN"
	ServerStatusOnline  ServerStatus = "ONLINE"
	ServerStatusOffline ServerStatus = "OFFLINE"
)

type Server struct {
	ServerID     string       `gorm:"primaryKey"`
	ServerName   string       `gorm:"not null;unique"`
	IPv4         string       `gorm:"not null;unique"`
	Status       ServerStatus `gorm:"not null;default:UNKNOWN"`
	Location     string
	OS           string
	IntervalTime int       `gorm:"not null;default:5"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	DeletedAt    gorm.DeletedAt
}
