package app

import (
	v1 "github.com/th1enq/ViettelSMS_ServerService/api/gen/go/server_service/v1"
	"github.com/th1enq/ViettelSMS_ServerService/internal/configs"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecases/server"
)

type App struct {
	Cfg            *configs.Config
	UC             server.UseCase
	UserGRPCServer v1.ServerServiceServer
}

func New(
	cfg *configs.Config,
	uc server.UseCase,
	userGRPCServer v1.ServerServiceServer,
) *App {
	return &App{
		Cfg:            cfg,
		UC:             uc,
		UserGRPCServer: userGRPCServer,
	}
}
