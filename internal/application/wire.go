//go:build wireinject
// +build wireinject

package application

import (
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/controller"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/presenter"
	log "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/logger"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/postgres"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/repository"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/service"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
)

func InitApp() (*Application, error) {
	panic(wire.Build(
		NewApplication,
		config.ConfigWireSet,
		log.LoggerSet,
		http.HTTPServerSet,
		server.UseCaseSet,
		controller.ControllerWireSet,
		repository.RepositorySet,
		postgres.PostgresWireSet,
		service.ExcelizeServiceSet,
		presenter.PresenterWireSet,
	))
}
