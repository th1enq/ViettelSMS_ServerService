package consumer

import (
	"context"
	"encoding/json"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
	"go.uber.org/zap"
)

type HandleFunc interface {
	Handle(ctx context.Context, topic string, payload []byte) error
}

type handleFunc struct {
	logger  *zap.Logger
	usecase server.UseCase
}

func NewHandlerFunc(
	logger *zap.Logger,
	usecase server.UseCase,
) HandleFunc {
	return &handleFunc{
		logger:  logger,
		usecase: usecase,
	}
}

var HandlerFunc = wire.NewSet(NewHandlerFunc)

func (h *handleFunc) Handle(ctx context.Context, topic string, payload []byte) error {
	h.logger.Info("Handling message", zap.String("topic", topic), zap.ByteString("payload", payload))
	var msg dto.UpdateStatusMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		h.logger.Error("failed to unmarshal payload", zap.Error(err))
		return err
	}

	return h.usecase.UpdateStatus(ctx, msg)
}
