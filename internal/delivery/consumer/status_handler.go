package consumer

import (
	"context"
	"encoding/json"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
	"go.uber.org/zap"
)

type StatusHandleFunc interface {
	Handle(ctx context.Context, topic string, payload []byte) error
}

type statusHandleFunc struct {
	logger  *zap.Logger
	usecase server.UseCase
}

func NewStatusHandlerFunc(
	logger *zap.Logger,
	usecase server.UseCase,
) StatusHandleFunc {
	return &statusHandleFunc{
		logger:  logger,
		usecase: usecase,
	}
}

func (h *statusHandleFunc) Handle(ctx context.Context, topic string, payload []byte) error {
	h.logger.Info("Handling message", zap.String("topic", topic), zap.ByteString("payload", payload))

	var msg dto.UpdateStatusMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		h.logger.Error("failed to unmarshal payload", zap.Error(err))
		return err
	}

	return h.usecase.UpdateStatus(ctx, msg)
}
