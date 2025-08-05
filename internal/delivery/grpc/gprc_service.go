package grpc

import (
	"ViettelSMS_ServerService/internal/commands"
	"ViettelSMS_ServerService/internal/service"
	"ViettelSMS_ServerService/proto/server"
	"context"

	"github.com/go-playground/validator"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

type gprcService struct {
	server.UnimplementedServerServiceServer
	logger        *zap.Logger
	validate      *validator.Validate
	serverService *service.ServerService
}

func NewGRPCService(logger *zap.Logger, serverService *service.ServerService, validate *validator.Validate) *gprcService {
	return &gprcService{
		logger:        logger,
		serverService: serverService,
		validate:      validate,
	}
}

func (g *gprcService) CreateServer(ctx context.Context, req *server.CreateServerRequest) (*server.CreateServerResponse, error) {
	g.logger.Info("CreateServer called", zap.Any("request", req))

	aggregateID := uuid.NewV4().String()
	command := commands.CreateServerCommand{
		AggregateID:  aggregateID,
		ServerID:     req.GetServerId(),
		ServerName:   req.GetServerName(),
		IPv4:         req.GetIpv4(),
		Location:     req.GetLocation(),
		OS:           req.GetOs(),
		IntervalTime: req.GetIntervalCheckTime(),
	}

	if err := g.validate.StructCtx(ctx, command); err != nil {
		g.logger.Error("Validation error", zap.Error(err))
		return nil, err
	}

	err := g.serverService.Commands.CreateServer.Handle(ctx, command)
	if err != nil {
		g.logger.Error("Error handling CreateServer command", zap.Error(err))
	}
	return &server.CreateServerResponse{
		AggregateId: aggregateID,
	}, nil
}
