package router

import (
	"context"

	"github.com/google/wire"
	v1 "github.com/th1enq/ViettelSMS_ServerService/api/gen/go/server_service/v1"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecases/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type serverGRPCServer struct {
	v1.UnimplementedServerServiceServer
	uc     server.UseCase
	logger *zap.Logger
}

var ServerGRPCServerSet = wire.NewSet(NewGRPCServerServer)

func NewGRPCServerServer(
	grpc *grpc.Server,
	uc server.UseCase,
	logger *zap.Logger,
) v1.ServerServiceServer {
	svc := serverGRPCServer{
		uc:     uc,
		logger: logger,
	}

	v1.RegisterServerServiceServer(grpc, &svc)

	reflection.Register(grpc)

	return &svc
}

func (s *serverGRPCServer) CreateServer(ctx context.Context, req *v1.CreateServerRequest) (*v1.CreateServerResponse, error) {
	s.logger.Info("CreateServer called", zap.Any("request", req))

	server := &domain.Server{
		ServerID:     req.GetServerId(),
		ServerName:   req.GetServerName(),
		IPv4:         req.GetIpv4(),
		Location:     req.GetLocation(),
		OS:           req.GetOs(),
		IntervalTime: req.GetIntervalCheckTime(),
	}
	if err := s.uc.CreateServer(ctx, server); err != nil {
		if err == domain.ErrServerExist {
			s.logger.Warn("Server already exists", zap.String("serverID", server.ServerID), zap.String("serverName", server.ServerName))
			return nil, status.Error(codes.AlreadyExists, "server already exists")
		} else {
			s.logger.Error("Failed to create server", zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create server")
		}
	}

	s.logger.Info("Server created successfully")
	return &v1.CreateServerResponse{
		Server: &v1.Server{
			ServerId:          server.ServerID,
			ServerName:        server.ServerName,
			Ipv4:              server.IPv4,
			Location:          server.Location,
			Os:                server.OS,
			IntervalCheckTime: server.IntervalTime,
		},
	}, nil
}
