package http

import (
	"github.com/gin-gonic/gin"
	server_usecase "github.com/th1enq/ViettelSMS_ServerService/internal/usecase/server"
	"go.uber.org/zap"
)

type Controller struct {
	usecase server_usecase.UseCase
	logger  *zap.Logger
}

func NewController(
	usecase server_usecase.UseCase,
	logger *zap.Logger,
) *Controller {
	return &Controller{
		usecase: usecase,
		logger:  logger,
	}
}

func (s *Controller) Create(c *gin.Context) {
	return
}
