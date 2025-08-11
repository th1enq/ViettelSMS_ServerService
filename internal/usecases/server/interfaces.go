package server

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
)

type UseCase interface {
	CreateServer(ctx context.Context, server *domain.Server) error
	UpdateServer(ctx context.Context, serverID string, update domain.UpdateServerParams) error
	DeleteServer(ctx context.Context, serverID string) error
	ViewServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) ([]*domain.Server, int, error)

	ImportServer(ctx context.Context, filePath string) (*domain.ImportServerResponse, error)
	ExportServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) (string, error)
}
