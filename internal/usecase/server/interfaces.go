package server

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
)

type UseCase interface {
	CreateServer(ctx context.Context, serverCreateRequest dto.CreateServerParams) (*dto.ServerResponse, error)
	UpdateServer(ctx context.Context, serverID string, update dto.UpdateServerParams) (*dto.ServerResponse, error)
	DeleteServer(ctx context.Context, serverID string) error
	ViewServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*dto.ServerResponse, int, error)

	ImportServer(ctx context.Context, filePath string) (*dto.ImportServerResponse, error)
	ExportServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) (string, error)

	UpdateStatus(ctx context.Context, updateStatus dto.UpdateStatusMessage) error
}
