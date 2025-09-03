package application

import (
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/consumer"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/controller"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/middleware"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/presenter"
	consumerGroup "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/consumer"
	log "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/logger"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/postgres"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/repository"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/service"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
)

func InitApp() (*Application, error) {
	config := config.LoadConfig()

	logger, err := log.LoadLogger(config)
	if err != nil {
		return nil, err
	}

	db, err := postgres.NewPostgresDB(config, logger)
	if err != nil {
		return nil, err
	}

	excelSrv := service.NewExcelizeService(logger)

	repo := repository.NewServerRepository(db)

	usecase := server.NewServerUseCase(
		repo,
		excelSrv,
		logger,
	)

	presenter := presenter.NewPresenter()
	middleware := middleware.NewJWTMiddleware(presenter, []byte(config.JWT.Secret))
	controller := controller.NewController(usecase, logger, presenter)

	httpServer := http.NewHttpServer(config, controller, middleware, logger)

	statusConsumer, err := consumerGroup.NewConsumer(
		config,
		logger,
		config.Consumer.StatusConsumer,
	)

	statusHandleFunc := consumer.NewStatusHandlerFunc(logger, usecase)

	if err != nil {
		return nil, err
	}

	rootConsumer := consumer.NewRoot(
		logger,
		statusConsumer,
		statusHandleFunc,
	)

	app := NewApplication(httpServer, rootConsumer, logger)
	return app, nil
}
