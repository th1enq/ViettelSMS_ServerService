package entity

import "time"

type ServerStatus string

const (
	ServerStatusUnknown ServerStatus = "UNKNOWN"
	ServerStatusOnline  ServerStatus = "ONLINE"
	ServerStatusOffline ServerStatus = "OFFLINE"
)

type Server struct {
	ServerID     string       `json:"server_id" gorm:"primaryKey;type:varchar(32);not null"`
	ServerName   string       `json:"server_name" gorm:"type:varchar(64);not null;unique"`
	Status       ServerStatus `json:"status" gorm:"type:varchar(16);not null;default:'UNKNOWN'"`
	IPv4         string       `json:"ipv4" gorm:"type:varchar(15);not null"`
	Location     string       `json:"location" gorm:"type:varchar(128)"`
	OS           string       `json:"os" gorm:"type:varchar(32)"`
	IntervalTime uint64       `json:"interval_time" gorm:"type:bigint;not null"`
	CreatedAt    time.Time    `json:"created_at" gorm:"type:timestamp;not null"`
	DeletedAt    time.Time    `json:"deleted_at" gorm:"type:timestamp;default:null"`
}
