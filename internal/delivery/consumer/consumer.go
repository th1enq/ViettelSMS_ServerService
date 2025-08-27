package consumer

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/consumer"
	"go.uber.org/zap"
)

const (
	UPTIME_TOPIC = "server_heartbeat"
)

type (
	Root interface {
		Start(ctx context.Context) error
	}

	root struct {
		logger   *zap.Logger
		consumer consumer.Consumer
		handler  HandleFunc
	}
)

func NewRoot(
	logger *zap.Logger,
	handler HandleFunc,
	consumer consumer.Consumer,
) Root {
	return &root{
		logger:   logger,
		consumer: consumer,
		handler:  handler,
	}
}

func (r *root) Start(ctx context.Context) error {
	r.logger.Info("Starting Kafka consumer...")

	r.consumer.RegisterHandler(
		UPTIME_TOPIC,
		func(ctx context.Context, queueName string, payload []byte) error {
			return r.handler.Handle(ctx, queueName, payload)
		},
	)

	r.logger.Info("Kafka consumer started, waiting for messages...")
	go func() {
		if err := r.consumer.Start(ctx); err != nil {
			r.logger.Error("Failed to start consumer", zap.Error(err))
		}
	}()
	return nil
}
