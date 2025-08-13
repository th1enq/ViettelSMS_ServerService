package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
	domain "github.com/th1enq/ViettelSMS_ServerService/internal/domain/errors"
	repo "github.com/th1enq/ViettelSMS_ServerService/internal/domain/repository"
	srv "github.com/th1enq/ViettelSMS_ServerService/internal/domain/service"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	NUMBER_OF_WORKERS = 15
	BATCH_SIZE        = 150
)

type serverUseCase struct {
	repo     repo.ServerRepository
	excelSrv srv.XLSXService
	logger   *zap.Logger
}

func NewServerUseCase(
	repo repo.ServerRepository,
	excelSrv srv.XLSXService,
	logger *zap.Logger,
) UseCase {
	return &serverUseCase{
		repo:     repo,
		excelSrv: excelSrv,
		logger:   logger,
	}
}

var UseCaseSet = wire.NewSet(NewServerUseCase)

func (s *serverUseCase) CreateServer(ctx context.Context, serverCreateRequest dto.CreateServerParams) error {
	s.logger.Info("CreateServer called", zap.Any("request", serverCreateRequest))

	if exist, err := s.repo.ExistByNameOrID(ctx, serverCreateRequest.ServerID, serverCreateRequest.ServerName); err != nil {
		s.logger.Error("failed to check server existence", zap.Error(err))
		return domain.ErrInternalServer
	} else if exist {
		s.logger.Warn("Server already exists", zap.String("server_id", serverCreateRequest.ServerID), zap.String("server_name", serverCreateRequest.ServerName))
		return domain.ErrServerExist
	}

	server := &entity.Server{
		ServerID:     serverCreateRequest.ServerID,
		ServerName:   serverCreateRequest.ServerName,
		IPv4:         serverCreateRequest.IPv4,
		IntervalTime: serverCreateRequest.IntervalTime,
	}
	if serverCreateRequest.Location != nil {
		server.Location = *serverCreateRequest.Location
	}
	if serverCreateRequest.OS != nil {
		server.OS = *serverCreateRequest.OS
	}

	if err := s.repo.Create(ctx, server); err != nil {
		s.logger.Error("failed to create server", zap.Error(err))
		return domain.ErrInternalServer
	}
	s.logger.Info("Server created successfully", zap.Any("server", server))
	return nil
}

func (s *serverUseCase) DeleteServer(ctx context.Context, serverID string) error {
	s.logger.Info("DeleteServer called", zap.Any("server_id", serverID))

	if _, err := s.repo.GetByField(ctx, "server_id", serverID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Server not found", zap.String("server_id", serverID))
			return domain.ErrServerNotFound
		}
		s.logger.Error("failed to get server by ID", zap.String("server_id", serverID), zap.Error(err))
		return domain.ErrInternalServer
	}

	if err := s.repo.Delete(ctx, serverID); err != nil {
		s.logger.Error("failed to delete server", zap.String("server_id", serverID))
		return domain.ErrInternalServer
	}

	s.logger.Info("Server delete successfully", zap.String("server_id", serverID))
	return nil
}

func (s *serverUseCase) ExportServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) (string, error) {
	s.logger.Info("ViewServer called", zap.Any("filter", filter), zap.Any("pagination", pagination))

	servers, _, err := s.ViewServer(ctx, filter, pagination)
	if err != nil {
		s.logger.Error("failed to get servers", zap.Error(err))
		return "", err
	}

	file := excelize.NewFile()
	streamWriter, err := file.NewStreamWriter("Sheet1")
	if err != nil {
		s.logger.Error("failed to create stream writer for export", zap.Error(err))
		return "", domain.ErrInternalServer
	}

	streamWriter.SetRow("A1", []interface{}{
		"server_id", "server_name", "IPv4", "status", "location", "os", "interval_time",
	})

	for rowIndex, server := range servers {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+2)
		if err := streamWriter.SetRow(cell, []interface{}{
			server.ServerID,
			server.ServerName,
			server.IPv4,
			server.Status,
			server.Location,
			server.OS,
			server.IntervalTime,
		}); err != nil {
			s.logger.Error("failed to write server data to export file", zap.Any("server", server), zap.Error(err))
			return "", domain.ErrInternalServer
		}
	}

	if err := streamWriter.Flush(); err != nil {
		s.logger.Error("failed to flush stream writer for export", zap.Error(err))
		return "", domain.ErrInternalServer
	}

	_ = os.MkdirAll("./exports", 0755)

	filePath := fmt.Sprintf("./exports/servers_%d.xlsx", time.Now().Unix())
	if err := file.SaveAs(filePath); err != nil {
		s.logger.Error("failed to save export file", zap.String("file_path", filePath), zap.Error(err))
		return "", domain.ErrInternalServer
	}

	s.logger.Info("Export file successfully", zap.String("file_path", filePath), zap.Int("total_server", len(servers)))
	return filePath, nil
}

func (s *serverUseCase) ImportServer(ctx context.Context, filePath string) (*dto.ImportServerResponse, error) {
	s.logger.Info("ImportServer called", zap.String("filePath", filePath))

	rows, err := s.excelSrv.GetRows(filePath)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidFile) {
			s.logger.Warn("invalid file format or content", zap.String("filePath", filePath))
			return nil, domain.ErrInvalidFile
		}
		return nil, domain.ErrInternalServer
	}

	if len(rows) <= 2 || s.excelSrv.Validate(rows[0]) != nil {
		s.logger.Warn("file import must contain at least 2 rows (header + data)")
		return nil, domain.ErrInvalidFile
	}

	result := dto.ImportServerResponse{
		SuccessCount:   0,
		SuccessServers: make([]string, 0),
		FailedCount:    0,
		FailedServers:  make([]string, 0),
	}

	allServers := make([]*entity.Server, 0)

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
		batchCopy := batches

		s.logger.Info("Processing batch", zap.Int("start", i), zap.Int("end", end))

		workerPool.Submit(func() {
			if err := s.repo.BatchCreate(ctx, batchCopy); err == nil {
				mu.Lock()
				result.SuccessCount += len(batchCopy)
				for _, server := range batchCopy {
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

func (s *serverUseCase) UpdateServer(ctx context.Context, serverID string, update dto.UpdateServerParams) error {
	s.logger.Info("UpdateServer called", zap.String("server_id", serverID), zap.Any("update", update))

	server, err := s.repo.GetByField(ctx, "server_id", serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Server not found", zap.String("server_id", serverID))
			return domain.ErrServerNotFound
		}
		s.logger.Error("failed to get server by ID", zap.String("server_id", serverID), zap.Error(err))
		return err
	}

	if update.ServerName != nil {
		exists, err := s.repo.GetByField(ctx, "server_name", update.ServerName)
		if err == nil && exists != nil && exists.ServerID != server.ServerID {
			s.logger.Warn("Server with the same name already exists", zap.String("server_name", *update.ServerName))
			return domain.ErrServerExist
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Error("failed to check server existence by server_name", zap.String("server_name", *update.ServerName), zap.Error(err))
			return err
		}
		server.ServerName = *update.ServerName
	}
	if update.IPv4 != nil {
		server.IPv4 = *update.IPv4
	}
	if update.Location != nil {
		server.Location = *update.Location
	}
	if update.OS != nil {
		server.OS = *update.OS
	}
	if update.IntervalTime != nil {
		server.IntervalTime = *update.IntervalTime
	}

	if err := s.repo.Update(ctx, server); err != nil {
		s.logger.Error("failed to update server", zap.Any("server", server), zap.Error(err))
		return err
	}
	s.logger.Info("Update server successfully", zap.Any("server", server))
	return nil
}

func (s *serverUseCase) ViewServer(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error) {
	s.logger.Info("ViewServer called", zap.Any("filter", filter), zap.Any("pagination", pagination))

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
		return nil, 0, domain.ErrInternalServer
	}
	s.logger.Info("Servers retrieved successfully", zap.Int("total", total), zap.Any("servers", servers))
	return servers, total, nil
}
