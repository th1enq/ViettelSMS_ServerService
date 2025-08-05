package commands

import (
	"ViettelSMS_ServerService/pkg/es"
	"context"

	"go.uber.org/zap"
)

type CreateServerCommand struct {
	AggregateID  string `json:"aggregate_id" validate:"required"`
	ServerID     string `json:"server_id" validate:"required"`
	ServerName   string `json:"server_name" validate:"required"`
	IPv4         string `json:"ipv4" validate:"required,ipv4"`
	Location     string `json:"location"`
	OS           string `json:"os"`
	IntervalTime uint32 `json:"interval_time" validate:"required,gte=5"`
}

type CreateServer interface {
	Handle(ctx context.Context, cmd CreateServerCommand) error
}

type createServerCmdHandler struct {
	logger         *zap.Logger
	aggregateStore es.AggregateStore
}

func NewCreateServerCmdHandler(logger *zap.Logger, aggregateStore es.AggregateStore) *createServerCmdHandler {
	return &createServerCmdHandler{
		logger:         logger,
		aggregateStore: aggregateStore,
	}

}

func (c *createServerCmdHandler) Handle(ctx context.Context, cmd CreateServerCommand) error {
}
