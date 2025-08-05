package domain

import "time"

type ServerStatus string

const (
	ServerStatusOn        ServerStatus = "ON"
	ServerStatusOff       ServerStatus = "OFF"
	ServerStatusUndefined ServerStatus = "UNDEFINED"
)

type Server struct {
	AggregateID  string       `gorm:"primaryKey"`
	ServerID     string       `gorm:"index;unique;not null"`
	ServerName   string       `gorm:"index;unique;not null"`
	Status       ServerStatus `gorm:"not null;default:'OFF'"`
	IPv4         string
	Location     string
	OS           string
	IntervalTime uint32    `gorm:"default:10"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func NewServer(aggregateID string) *Server {
	return &Server{
		AggregateID: aggregateID,
		Status:      ServerStatusUndefined,
	}
}
