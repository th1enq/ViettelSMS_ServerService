package service

import (
	"ViettelSMS_ServerService/internal/commands"
	"ViettelSMS_ServerService/pkg/es"

	"go.uber.org/zap"
)

type ServerService struct {
	Commands *commands.ServerCommands
	logger   *zap.Logger
}

func NewServerService(
	logger *zap.Logger,
	aggregateStore es.AggregateStore,
) *ServerService {
	serverCommands := commands.NewServerCommands(
		commands.NewCreateServerCmdHandler(logger, aggregateStore),
	)
	return &ServerService{
		Commands: serverCommands,
		logger:   logger,
	}
}
