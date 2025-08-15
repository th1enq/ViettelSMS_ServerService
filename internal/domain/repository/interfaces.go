package repo

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
)

type ServerRepository interface {
	ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error)
	Create(ctx context.Context, server *entity.Server) error
	Delete(ctx context.Context, serverID string) error
	GetByField(ctx context.Context, field string, value interface{}) (*entity.Server, error)
	Update(ctx context.Context, server *entity.Server) error
	GetServers(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error)
	BatchCreate(ctx context.Context, servers []*entity.Server) ([]*string, error)
}
