package application

import (
	"context"
	"syscall"

	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/consumer"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http"
	"github.com/th1enq/ViettelSMS_ServerService/internal/utils"
	"go.uber.org/zap"
)

type Application struct {
	httpServer   http.Server
	rootConsumer consumer.Root
	logger       *zap.Logger
}

func NewApplication(
	httpServer http.Server,
	rootConsumer consumer.Root,
	logger *zap.Logger,
) *Application {
	return &Application{
		httpServer:   httpServer,
		rootConsumer: rootConsumer,
		logger:       logger,
	}
}

func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting application ...")

	app.logger.Info("Starting HTTP Server ...")
	go func() {
		if err := app.httpServer.Start(ctx); err != nil {
			app.logger.Error("HTTP Server failed to start", zap.Error(err))
		}
	}()

	app.logger.Info("Starting Kafka Consumer ...")
	go func() {
		if err := app.rootConsumer.Start(ctx); err != nil {
			app.logger.Error("Kafka Consumer failed to start", zap.Error(err))
		}
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)

	return nil
}
