package postgres

import (
	"fmt"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type postgresDB struct {
	db *gorm.DB
}

func NewPostgresDB(cfg *config.Config, logger *zap.Logger) (DBEngine, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.DBName,
		cfg.Postgres.Port)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Postgres.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxOpenConns)

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Postgres.Host),
		zap.Int("port", cfg.Postgres.Port),
		zap.String("database", cfg.Postgres.DBName),
	)
	return &postgresDB{db: gormDB}, nil
}

func (p *postgresDB) GetDB() *gorm.DB {
	return p.db
}

var PostgresWireSet = wire.NewSet(NewPostgresDB)
