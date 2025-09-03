package dto

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
)

type (
	CreateServerParams struct {
		ServerID     string  `json:"server_id" binding:"required"`
		ServerName   string  `json:"server_name" binding:"required"`
		IPv4         string  `json:"ipv4" binding:"required,ipv4"`
		Location     *string `json:"location"`
		OS           *string `json:"os"`
		IntervalTime int     `json:"interval_time" binding:"required,min=1,max=60"`
	}

	UpdateServerParams struct {
		ServerName   *string `json:"server_name"`
		IPv4         *string `json:"ipv4"`
		Location     *string `json:"location"`
		OS           *string `json:"os"`
		IntervalTime *int    `json:"interval_time"`
	}

	ServerFilterOptions struct {
		ServerName *string              `form:"server_name"`
		Status     *entity.ServerStatus `form:"status" binding:"omitempty,oneof=ONLINE OFFLINE UNKNOWN"`
	}

	ServerPaginationOptions struct {
		Page      int    `form:"page" binding:"min=1" default:"1"`
		PageSize  int    `form:"page_size" binding:"min=1,max=100" default:"10"`
		SortBy    string `form:"sort_by" binding:"omitempty,oneof=server_name ipv4 status location os interval_time" default:"server_name"`
		SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc" default:"asc"`
	}

	ImportServerResponse struct {
		SuccessCount   int      `json:"success_count"`
		SuccessServers []string `json:"server_ids"`
		FailedCount    int      `json:"failed_count"`
		FailedServers  []string `json:"failed_servers"`
	}

	ServerResponse struct {
		ServerID     string              `json:"server_id"`
		ServerName   string              `json:"server_name"`
		IPv4         string              `json:"ipv4"`
		Status       entity.ServerStatus `json:"status"`
		Location     string              `json:"location"`
		OS           string              `json:"os"`
		IntervalTime int                 `json:"interval_time"`
	}

	UpdateStatusMessage struct {
		ServerID  string              `json:"server_id"`
		Status    entity.ServerStatus `json:"status"`
		Timestamp time.Time           `json:"timestamp"`
	}

	Claims struct {
		Sub     uint     `json:"sub"`
		Scopes  []string `json:"scopes"`
		Blocked bool     `json:"blocked"`
		jwt.RegisteredClaims
	}
)

func ToServerResponse(server *entity.Server) *ServerResponse {
	return &ServerResponse{
		ServerID:     server.ServerID,
		ServerName:   server.ServerName,
		IPv4:         server.IPv4,
		Status:       server.Status,
		Location:     server.Location,
		OS:           server.OS,
		IntervalTime: server.IntervalTime,
	}
}

func ToServersResponse(servers []*entity.Server) []*ServerResponse {
	responses := make([]*ServerResponse, len(servers))
	for i, server := range servers {
		responses[i] = ToServerResponse(server)
	}
	return responses
}
