package repository

import (
	"context"
	"fmt"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
	repo "github.com/th1enq/ViettelSMS_ServerService/internal/domain/repository"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/postgres"
	"gorm.io/gorm/clause"
)

type ServerRepository struct {
	db postgres.DBEngine
}

var RepositorySet = wire.NewSet(NewServerRepository)

func NewServerRepository(db postgres.DBEngine) repo.ServerRepository {
	return &ServerRepository{db: db}
}

func (s *ServerRepository) Create(ctx context.Context, server *entity.Server) error {
	return s.db.GetDB().WithContext(ctx).Create(server).Error
}

func (s *ServerRepository) Delete(ctx context.Context, serverID string) error {
	return s.db.GetDB().WithContext(ctx).Where("server_id = ?", serverID).Delete(&entity.Server{}).Error
}

func (s *ServerRepository) ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error) {
	var count int64
	err := s.db.GetDB().WithContext(ctx).Model(&entity.Server{}).Where("server_id = ? OR server_name = ?", serverID, serverName).Count(&count).Error
	return count > 0, err
}

func (s *ServerRepository) GetByField(ctx context.Context, field string, value interface{}) (*entity.Server, error) {
	var server entity.Server
	err := s.db.GetDB().WithContext(ctx).Model(&entity.Server{}).Where(field+" = ?", value).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (s *ServerRepository) Update(ctx context.Context, server *entity.Server) error {
	return s.db.GetDB().WithContext(ctx).Save(server).Error
}

func (s *ServerRepository) GetServers(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error) {
	var servers []*entity.Server
	var total int64

	query := s.db.GetDB().WithContext(ctx).Model(&entity.Server{})

	if filter.ServerName != nil {
		query = query.Where("server_name LIKE ?", "%"+*filter.ServerName+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := fmt.Sprintf("%s %s", pagination.SortBy, pagination.SortOrder)

	if err := query.Order(orderBy).
		Offset((pagination.Page - 1) * pagination.PageSize).
		Limit(pagination.PageSize).
		Find(&servers).Error; err != nil {
		return nil, 0, err
	}

	return servers, int(total), nil
}

func (s *ServerRepository) BatchCreate(ctx context.Context, servers []*entity.Server) error {
	if err := s.db.GetDB().WithContext(ctx).
		Clauses(
			clause.OnConflict{
				DoNothing: true,
			}).
		Clauses(
			clause.Returning{},
		).
		Create(&servers).Error; err != nil {
		return err
	}

	return nil
}
