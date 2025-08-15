package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/controller"
	"go.uber.org/zap"
)

type (
	Server interface {
		Start(ctx context.Context) error
	}

	server struct {
		config     *config.Config
		controller *controller.Controller
		logger     *zap.Logger
	}
)

func NewHttpServer(
	config *config.Config,
	controller *controller.Controller,
	logger *zap.Logger,
) Server {
	return &server{
		config:     config,
		controller: controller,
		logger:     logger,
	}
}

var HTTPServerSet = wire.NewSet(NewHttpServer)

func (s *server) RegisterRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")

	server := v1.Group("/server")
	{
		server.POST("/", s.controller.Create)
		server.DELETE("/:id", s.controller.Delete)
		server.PUT("/:id", s.controller.Update)
		server.GET("/", s.controller.View)

		server.POST("/import", s.controller.Import)
		server.GET("/export", s.controller.Export)
	}

	return router
}

func (s *server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler: s.RegisterRoutes(),
	}
	s.logger.Info("HTTP server starting", zap.String("host", s.config.Postgres.Host), zap.Int("port", s.config.Server.Port))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
	return nil
}
