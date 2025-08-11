//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/app/router"
	"github.com/th1enq/ViettelSMS_ServerService/internal/configs"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/repository"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecases/server"
	"github.com/th1enq/ViettelSMS_ServerService/pkg/log"
	"github.com/th1enq/ViettelSMS_ServerService/pkg/postgres"
	"google.golang.org/grpc"
)

func InitApp(
	cfg *configs.Config,
	grpcServer *grpc.Server,
) (*App, error) {
	panic(wire.Build(
		New,
		router.ServerGRPCServerSet,
		repository.RepositorySet,
		server.UseCaseSet,
		postgres.PostgresSet,
		log.LoggerSet,
		domain.ExcelizeServiceSet,
	))
}
