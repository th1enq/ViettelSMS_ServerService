package postgres_repository

import (
	"ViettelSMS_ServerService/internal/domain"
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewPostgresRepository(db *gorm.DB, logger *zap.Logger) domain.PostgresRepository {
	return &serverPostgresRepository{
		logger: logger,
		db:     db,
	}
}

type serverPostgresRepository struct {
	logger *zap.Logger
	db     *gorm.DB
}

func (s *serverPostgresRepository) CreateServer(ctx context.Context, server *domain.Server) error {
	return s.db.WithContext(ctx).Create(server).Error
}

func (s *serverPostgresRepository) DeleteServer(ctx context.Context, server_id string) error {
	return s.db.WithContext(ctx).Where("server_id = ?", server_id).Delete(&domain.Server{}).Error
}

func (s *serverPostgresRepository) UpdateServer(ctx context.Context, server *domain.Server) error {
	return s.db.WithContext(ctx).Save(server).Error
}

func (s *serverPostgresRepository) UpdateStatus(ctx context.Context, server_id string, status string) error {
	return s.db.WithContext(ctx).Model(&domain.Server{}).
		Where("server_id = ?", server_id).
		Update("status", status).Error
}
