package models

import (
	"ViettelSMS_ServerService/proto/server_service"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	ServerID     string    `json:"server_id"`
	ServerName   string    `json:"server_name"`
	IPv4         string    `json:"ipv4"`
	Location     string    `json:"location"`
	OS           string    `json:"os"`
	IntervalTime int64     `json:"interval_time"`
	CreatedAt    time.Time `json:"created_at"`
}

func ServerToGrpcMessage(server *Server) *server_service.Server {
	return &server_service.Server{
		ServerId:     server.ServerID,
		ServerName:   server.ServerName,
		Ipv4:         server.IPv4,
		Location:     server.Location,
		Os:           server.OS,
		IntervalTime: server.IntervalTime,
		CreatedAt:    timestamppb.New(server.CreatedAt),
	}
}
