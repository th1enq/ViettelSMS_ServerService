package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	NUMBER_OF_WORKERS = 15
	BATCH_SIZE        = 150
)

type service struct {
	repo     domain.ServerRepository
	excelSrv domain.XLSXService
	logger   *zap.Logger
}

var UseCaseSet = wire.NewSet(NewService)

func NewService(
	repo domain.ServerRepository,
	excelSrv domain.XLSXService,
	logger *zap.Logger,
) UseCase {
	return &service{repo: repo, excelSrv: excelSrv, logger: logger}
}

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

func (s *service) DeleteServer(ctx context.Context, serverID string) error {
	s.logger.Info("DeleteServer called", zap.String("serverID", serverID))
	if _, err := s.repo.GetByField(ctx, "server_id", serverID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Server not found", zap.String("serverID", serverID))
			return domain.ErrServerNotFound
		}
		s.logger.Error("failed to get server by ID", zap.String("serverID", serverID), zap.Error(err))
		return err
	}
	if err := s.repo.Delete(ctx, serverID); err != nil {
		s.logger.Error("failed to delete server", zap.Error(err))
		return err
	}
	s.logger.Info("Server deleted successfully", zap.String("serverID", serverID))
	return nil
}

// UpdateServer implements UseCase.
func (s *service) UpdateServer(ctx context.Context, serverID string, update domain.UpdateServerParams) error {
	s.logger.Info("UpdateServer called", zap.String("serverID", serverID), zap.Any("update", update))
	server, err := s.repo.GetByField(ctx, "server_id", serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Server not found", zap.String("serverID", serverID))
			return domain.ErrServerNotFound
		}
		s.logger.Error("failed to get server by ID", zap.String("serverID", serverID), zap.Error(err))
		return err
	}
	if update.ServerName != "" {
		exists, err := s.repo.GetByField(ctx, "server_name", update.ServerName)
		if err == nil && exists.ServerID != server.ServerID {
			s.logger.Warn("Server with the same name already exists", zap.String("serverName", update.ServerName))
			return domain.ErrServerExist
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to check server existence by name", zap.String("serverName", update.ServerName), zap.Error(err))
			return err
		}
		server.ServerName = update.ServerName
	}
	if update.IPv4 != "" {
		server.IPv4 = update.IPv4
	}
	if update.Location != "" {
		server.Location = update.Location
	}
	if update.OS != "" {
		server.OS = update.OS
	}
	if update.IntervalTime > 0 {
		server.IntervalTime = update.IntervalTime
	}

	if err := s.repo.Update(ctx, server); err != nil {
		s.logger.Error("failed to update server", zap.Error(err))
		return err
	}
	s.logger.Info("Server updated successfully", zap.Any("server", server))
	return nil
}

func (s *service) ViewServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) ([]*domain.Server, int, error) {
	s.logger.Info("ExportServer called", zap.Any("filter", filter), zap.Any("pagination", pagination))
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}
	if pagination.SortBy == "" {
		pagination.SortBy = "server_name"
	}
	if pagination.SortOrder == "" {
		pagination.SortOrder = "asc"
	}

	servers, total, err := s.repo.GetServers(ctx, filter, pagination)
	if err != nil {
		s.logger.Error("failed to get servers", zap.Error(err))
		return nil, 0, err
	}
	s.logger.Info("Servers retrieved successfully", zap.Int("total", total), zap.Any("servers", servers))
	return servers, total, nil
}

// ExportServer implements UseCase.
func (s *service) ExportServer(ctx context.Context, filter domain.ServerFilterOptions, pagination domain.ServerPaginationOptions) (string, error) {
	panic("unimplemented")
}

func (s *service) ImportServer(ctx context.Context, filePath string) (*domain.ImportServerResponse, error) {
	s.logger.Info("ImportServer called", zap.String("filePath", filePath))

	rows, err := s.excelSrv.GetRows(filePath)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFile) {
			s.logger.Warn("invalid file format or content", zap.String("filePath", filePath))
			return nil, domain.ErrInvalidFile
		}
		return nil, err
	}

	if len(rows) <= 2 || s.excelSrv.Validate(rows[0]) != nil {
		s.logger.Warn("file import must contain at least 2 rows (header + data)")
		return nil, domain.ErrInvalidFile
	}

	result := domain.ImportServerResponse{
		SuccessCount:   0,
		SuccessServers: make([]string, 0),
		FailedCount:    0,
		FailedServers:  make([]string, 0),
	}

	allServers := make([]*domain.Server, 0)

	for i := 1; i < len(rows); i++ {
		row := rows[i]

		server, err := s.excelSrv.Parse(row)
		if err != nil {
			result.FailedCount++
			result.FailedServers = append(result.FailedServers, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}
		allServers = append(allServers, server)
	}

	workerPool := workerpool.New(NUMBER_OF_WORKERS)
	var mu sync.Mutex
	successID := make(map[string]bool)

	for i := 0; i < len(allServers); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(allServers) {
			end = len(allServers)
		}

		batches := allServers[i:end]

		s.logger.Info("Processing batch", zap.Int("start", i), zap.Int("end", end))

		workerPool.Submit(func() {
			if err := s.repo.BatchCreate(ctx, batches); err == nil {
				mu.Lock()
				result.SuccessCount += len(batches)
				for _, server := range batches {
					successID[server.ServerID] = true
					result.SuccessServers = append(result.SuccessServers, fmt.Sprintf("Server ID: %s, Name: %s", server.ServerID, server.ServerName))
				}
				mu.Unlock()
			}
		})
	}
	workerPool.StopWait()

	result.FailedCount += len(allServers) - result.SuccessCount
	for _, server := range allServers {
		if !successID[server.ServerID] {
			result.FailedServers = append(result.FailedServers, fmt.Sprintf("Existing Server ID: %s, Name: %s", server.ServerID, server.ServerName))
		}
	}

	s.logger.Info("ImportServer completed", zap.Int("successCount", result.SuccessCount), zap.Int("failedCount", result.FailedCount))
	return &result, nil
}
