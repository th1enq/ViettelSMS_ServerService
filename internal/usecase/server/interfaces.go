package server

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
)

type UseCase interface {
	CreateServer(ctx context.Context, serverCreateRequest dto.CreateServerParams) error
	UpdateServer(ctx context.Context, serverID string, update dto.UpdateServerParams) error
	DeleteServer(ctx context.Context, serverID string) error
	ViewServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error)

	ImportServer(ctx context.Context, filePath string) (*dto.ImportServerResponse, error)
	ExportServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) (string, error)
}
