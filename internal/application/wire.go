//go:build wireinject
// +build wireinject

package application

import (
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http"
	log "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/logger"
)

func InitApp() (*Application, error) {
	panic(wire.Build(
		NewApplication,
		config.ConfigWireSet,
		log.LoggerSet,
		http.HTTPServerSet,
	))
}
