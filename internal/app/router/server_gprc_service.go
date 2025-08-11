package router

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/google/wire"
	v1 "github.com/th1enq/ViettelSMS_ServerService/api/gen/go/server_service/v1"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecases/server"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/api/httpbody"
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
		Message: "Server created successfully",
	}, nil
}

func (s *serverGRPCServer) DeleteServer(ctx context.Context, req *v1.DeleteServerRequest) (*v1.DeleteServerResponse, error) {
	s.logger.Info("DeleteServer called", zap.String("serverID", req.GetServerId()))

	if err := s.uc.DeleteServer(ctx, req.GetServerId()); err != nil {
		if errors.Is(err, domain.ErrServerNotFound) {
			s.logger.Warn("Server not found", zap.String("serverID", req.GetServerId()))
			return nil, status.Error(codes.NotFound, "server not found")
		}
		s.logger.Error("Failed to delete server", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete server")
	}

	s.logger.Info("Server deleted successfully", zap.String("serverID", req.GetServerId()))
	return &v1.DeleteServerResponse{
		Message: "Server deleted successfully",
	}, nil
}

func (s *serverGRPCServer) UpdateServer(ctx context.Context, req *v1.UpdateServerRequest) (*v1.UpdateServerResponse, error) {
	s.logger.Info("UpdateServer called", zap.Any("request", req))

	serverID := req.GetServerId()
	update := domain.UpdateServerParams{
		ServerName:   req.GetServerName(),
		IPv4:         req.GetIpv4(),
		Location:     req.GetLocation(),
		OS:           req.GetOs(),
		IntervalTime: req.GetIntervalCheckTime(),
	}

	if err := s.uc.UpdateServer(ctx, serverID, update); err != nil {
		if err == domain.ErrServerNotFound {
			s.logger.Warn("Server not found", zap.String("serverID", serverID))
			return nil, status.Error(codes.NotFound, "server not found")
		}
		s.logger.Error("Failed to update server", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update server")
	}

	s.logger.Info("Server updated successfully", zap.String("serverID", serverID))
	return &v1.UpdateServerResponse{
		Message: "Server updated successfully",
	}, nil
}

func (s *serverGRPCServer) ViewServer(ctx context.Context, req *v1.ViewServerRequest) (*v1.ViewServerResponse, error) {
	s.logger.Info("ViewServer called", zap.Any("request", req))

	filter := domain.ServerFilterOptions{
		ServerName: req.GetFilter().GetServerName(),
		Status:     domain.ServerStatus(req.GetFilter().GetStatus().String()),
	}

	pagination := domain.ServerPaginationOptions{
		Page:      int(req.GetPagination().GetPage()),
		PageSize:  int(req.GetPagination().GetPageSize()),
		SortBy:    req.GetPagination().GetSortBy(),
		SortOrder: req.GetPagination().GetSortOrder(),
	}

	servers, total, err := s.uc.ViewServer(ctx, filter, pagination)
	if err != nil {
		s.logger.Error("Failed to view server", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to view server")
	}

	s.logger.Info("Servers retrieved successfully", zap.Int("total", total), zap.Any("servers", servers))
	return &v1.ViewServerResponse{
		Servers:    domain.DomainToArrayProto(servers),
		TotalCount: int32(total),
		Message:    "Servers retrieved successfully",
	}, nil
}

func (s *serverGRPCServer) ImportServer(ctx context.Context, req *httpbody.HttpBody) (*v1.ImportServerResponse, error) {
	s.logger.Info("ImportServer called", zap.Any("request", req))
	if req.ContentType != "text/xlsx" {
		s.logger.Warn("Invalid content type for import", zap.String("contentType", req.ContentType))
		return nil, status.Error(codes.InvalidArgument, "invalid content type")
	}

	fileName := fmt.Sprintf("%s_import.xlsx", uuid.New().String())
	filePath := fmt.Sprintf("/tmp/%s", fileName)

	if err := os.WriteFile(filePath, req.Data, 0644); err != nil {
		s.logger.Error("Failed to write file", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to write file")
	}

	result, err := s.uc.ImportServer(ctx, filePath)
	if err != nil {
		s.logger.Error("Failed to import server", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to import server")
	}

	return &v1.ImportServerResponse{
		SuccessCount: uint32(result.SuccessCount),
		FailureCount: uint32(result.FailedCount),
		SuccessIds:   result.SuccessServers,
		FailureIds:   result.FailedServers,
		Message:      "File imported successfully",
	}, nil
}
