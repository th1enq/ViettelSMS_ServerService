package server

import (
	"context"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"go.uber.org/zap"
)

type service struct {
	repo   domain.ServerRepository
	logger *zap.Logger
}

var UseCaseSet = wire.NewSet(NewService)

func NewService(
	repo domain.ServerRepository,
	logger *zap.Logger,
) UseCase {
	return &service{repo: repo, logger: logger}
}

// CreateServer implements UseCase.
func (s *service) CreateServer(ctx context.Context, server *domain.Server) error {
	s.logger.Info("CreateServer called", zap.Any("server", server))

	if exist, err := s.repo.ExistByNameOrID(ctx, server.ServerID, server.ServerName); err != nil {
		s.logger.Error("failed to check server existence", zap.Error(err))
		return err
	} else if exist {
		s.logger.Warn("Server already exists", zap.String("serverID", server.ServerID), zap.String("serverName", server.ServerName))
		return domain.ErrServerExist
	}

	if err := s.repo.Create(ctx, server); err != nil {
		s.logger.Error("failed to create server", zap.Error(err))
		return err
	}
	s.logger.Info("Server created successfully", zap.Any("server", server))
	return nil
}

// DeleteServer implements UseCase.
func (s *service) DeleteServer(ctx context.Context, serverID string) error {
	panic("unimplemented")
}

// ExportServer implements UseCase.
func (s *service) ExportServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) (string, error) {
	panic("unimplemented")
}

// ImportServer implements UseCase.
func (s *service) ImportServer(ctx context.Context, filePath string) ([]domain.ImportServerResponse, error) {
	panic("unimplemented")
}

// UpdateServer implements UseCase.
func (s *service) UpdateServer(ctx context.Context, server *domain.Server) error {
	panic("unimplemented")
}

// ViewServer implements UseCase.
func (s *service) ViewServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) ([]domain.Server, int, error) {
	panic("unimplemented")
}
