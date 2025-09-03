package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/controller"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/middleware"
	"go.uber.org/zap"
)

type (
	Server interface {
		Start(ctx context.Context) error
	}

	server struct {
		config     *config.Config
		controller *controller.Controller
		middleware middleware.JWTMiddleware
		logger     *zap.Logger
	}
)

func NewHttpServer(
	config *config.Config,
	controller *controller.Controller,
	middleware middleware.JWTMiddleware,
	logger *zap.Logger,
) Server {
	return &server{
		config:     config,
		controller: controller,
		middleware: middleware,
		logger:     logger,
	}
}

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
	server.Use(s.middleware.RequireAuth())
	{
		server.POST("/", s.controller.Create)
		server.DELETE("/:id", s.middleware.RequireScope("server:delete"), s.controller.Delete)
		server.PUT("/:id", s.middleware.RequireScope("server:update"), s.controller.Update)
		server.GET("/", s.middleware.RequireScope("server:view"), s.controller.View)

		server.POST("/import", s.middleware.RequireScope("server:import"), s.controller.Import)
		server.GET("/export", s.middleware.RequireScope("server:export"), s.controller.Export)
	}

	return router
}

func (s *server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler: s.RegisterRoutes(),
	}
	s.logger.Info("HTTP server starting", zap.String("host", s.config.Server.Host), zap.Int("port", s.config.Server.Port))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
	return nil
}
