package dto

import "github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"

type (
	CreateServerParams struct {
		ServerID    string `json:"server_id"`
		ServerName  string `json:"server_name"`
		IPv4        string `json:"ipv4"`
		Location    string `json:"location"`
		OS          string `json:"os"`
		IntevalTime int64  `json:"interval_time"`
	}

	UpdateServerParams struct {
		ServerID     string `json:"server_id"`
		ServerName   string `json:"server_name"`
		IPv4         string `json:"ipv4"`
		Location     string `json:"location"`
		OS           string `json:"os"`
		IntervalTime uint64 `json:"interval_time"`
	}

	ServerFilterOptions struct {
		ServerName string              `json:"server_name"`
		Status     entity.ServerStatus `json:"status"`
	}

	ServerPaginationOptions struct {
		Page      int    `json:"page"`
		PageSize  int    `json:"page_size"`
		SortBy    string `json:"sort_by"`
		SortOrder string `json:"sort_order"`
	}

	ImportServerResponse struct {
		SuccessCount   int      `json:"success_count"`
		SuccessServers []string `json:"server_ids"`
		FailedCount    int      `json:"failed_count"`
		FailedServers  []string `json:"failed_servers"`
	}
)
