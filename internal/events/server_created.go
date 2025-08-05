package events

import "ViettelSMS_ServerService/pkg/es"

const (
	ServerCreatedEventType es.EventType = "SERVER_CREATED"
)

type ServerCreated struct {
	ServerID     string `json:"server_id"`
	ServerName   string `json:"server_name"`
	IPV4         string `json:"ipv4"`
	Location     string `json:"location"`
	OS           string `json:"os"`
	IntervalTime uint32 `json:"interval_time"`
}
