package main

import (
	"fmt"
	"net"
	"syscall"

	"github.com/th1enq/ViettelSMS_ServerService/internal/app"
	"github.com/th1enq/ViettelSMS_ServerService/internal/configs"
	"github.com/th1enq/ViettelSMS_ServerService/internal/utils"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := configs.NewConfig(configs.ConfigFilePath("configs/config.yaml"))
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	server := grpc.NewServer()

	_, err = app.InitApp(&cfg, server)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", cfg.ServerService.Host, cfg.ServerService.Port),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on %s:%d: %v", cfg.ServerService.Host, cfg.ServerService.Port, err))
	}
	defer func() {
		if err := l.Close(); err != nil {
			panic(fmt.Sprintf("Failed to close listener: %v", err))
		}
	}()

	err = server.Serve(l)
	if err != nil {
		panic(fmt.Sprintf("Failed to start gRPC server: %v", err))
	}

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
}
