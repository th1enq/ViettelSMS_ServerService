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
	ServerName   string       `gorm:"not null;index;unique"`
	IPv4         string       `gorm:"not null;unique"`
	Status       ServerStatus `gorm:"not null;default:UNKNOWN"`
	IntervalTime int          `gorm:"not null;default:5"`
	Location     string
	OS           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
