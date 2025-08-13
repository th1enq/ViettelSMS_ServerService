package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"go.uber.org/zap"
)

type (
	ServerInterface interface {
		Start(ctx context.Context) error
	}

	server struct {
		config *config.Config
		logger *zap.Logger
	}
)

func NewHttpServer(
	config *config.Config,
	logger *zap.Logger,
) ServerInterface {
	return &server{
		config: config,
		logger: logger,
	}
}

var HTTPServerSet = wire.NewSet(NewHttpServer)

func RegisterRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return nil
}

func (s *server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler: RegisterRoutes(),
	}
	s.logger.Info("HTTP server starting", zap.String("host", s.config.Postgres.Host), zap.Int("port", s.config.Server.Port))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("failed to start HTTP server", zap.Error(err))
	}
	return nil
}
