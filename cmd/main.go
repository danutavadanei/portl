package main

import (
	"github.com/danutavadanei/portl/broker"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/danutavadanei/portl/common"
	"github.com/danutavadanei/portl/config"
	"github.com/danutavadanei/portl/http"
	"github.com/danutavadanei/portl/ssh"

	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfig()

	logger, err := common.NewLogger(cfg)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	brokers := broker.NewStore()

	httpSrv := http.NewServer(logger, brokers, cfg)

	go func() {
		if err := httpSrv.Serve(); err != nil {
			logger.Error("Failed to run HTTP server",
				zap.Error(err),
			)
		}
	}()

	sshServ, err := ssh.NewServer(logger, brokers, cfg)
	if err != nil {
		logger.Error("failed to create SSH server", zap.Error(err))
	}

	go func() {
		if err := sshServ.Serve(); err != nil {
			logger.Error("Failed to run SSH server",
				zap.Error(err),
			)
		}
	}()

	<-sigChannel
}
