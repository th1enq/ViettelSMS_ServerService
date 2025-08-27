package application

import (
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/consumer"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/controller"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/presenter"
	consumer2 "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/consumer"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/producer"
	log "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/logger"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/postgres"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/repository"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/service"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
)

func InitApp() (*Application, error) {
	configConfig := config.LoadConfig()
	logger, err := log.LoadLogger(configConfig)
	if err != nil {
		return nil, err
	}
	dbEngine, err := postgres.NewPostgresDB(configConfig, logger)
	if err != nil {
		return nil, err
	}
	serverRepository := repository.NewServerRepository(dbEngine)
	xlsxService := service.NewExcelizeService(logger)
	useCase := server.NewServerUseCase(serverRepository, xlsxService, logger)
	presenterPresenter := presenter.NewPresenter()
	controllerController := controller.NewController(useCase, logger, presenterPresenter)
	httpServer := http.NewHttpServer(configConfig, controllerController, logger)
	handleFunc := consumer.NewHandlerFunc(logger, useCase)
	messageBroker, err := producer.NewBroker(configConfig, logger)
	if err != nil {
		return nil, err
	}
	consumerConsumer, err := consumer2.NewConsumer(configConfig, messageBroker, logger)
	if err != nil {
		return nil, err
	}
	root := consumer.NewRoot(logger, handleFunc, consumerConsumer)
	application := NewApplication(httpServer, root, logger)
	return application, nil
}
