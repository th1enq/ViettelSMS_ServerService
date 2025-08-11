package repository

import (
	"context"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"github.com/th1enq/ViettelSMS_ServerService/pkg/postgres"
)

type ServerRepository struct {
	db postgres.DBEngine
}

var RepositorySet = wire.NewSet(NewServerRepository)

// Create implements domain.ServerRepository.
func (s *ServerRepository) Create(ctx context.Context, server *domain.Server) error {
	panic("unimplemented")
}

// ExistByNameOrID implements domain.ServerRepository.
func (s *ServerRepository) ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error) {
	panic("unimplemented")
}

func NewServerRepository(db postgres.DBEngine) domain.ServerRepository {
	return &ServerRepository{db: db}
}
