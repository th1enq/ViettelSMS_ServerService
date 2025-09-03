package consumer

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/consumer"
	"go.uber.org/zap"
)

const (
	STATUS_UPDATE_TOPIC = "status_update"
)

type (
	Root interface {
		Start(ctx context.Context) error
	}

	root struct {
		logger            *zap.Logger
		statusConsumer    consumer.Consumer
		statusHandlerFunc StatusHandleFunc
	}
)

func NewRoot(
	logger *zap.Logger,
	statusConsumer consumer.Consumer,
	statusHandlerFunc StatusHandleFunc,
) Root {
	return &root{
		logger:            logger,
		statusConsumer:    statusConsumer,
		statusHandlerFunc: statusHandlerFunc,
	}
}

func (r *root) Start(ctx context.Context) error {
	r.logger.Info("Starting Kafka consumer...")

	r.statusConsumer.RegisterHandler(
		STATUS_UPDATE_TOPIC,
		func(ctx context.Context, queueName string, payload []byte) error {
			return r.statusHandlerFunc.Handle(ctx, queueName, payload)
		},
	)

	r.logger.Info("Kafka consumer started, waiting for messages...")
	go func() {
		if err := r.statusConsumer.Start(ctx); err != nil {
			r.logger.Error("Failed to start consumer", zap.Error(err))
		}
	}()
	return nil
}
