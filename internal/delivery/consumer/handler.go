package consumer

import (
	"context"

	"github.com/google/wire"
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

var cnt = 0

func (h *handleFunc) Handle(ctx context.Context, topic string, payload []byte) error {

	return nil
}
