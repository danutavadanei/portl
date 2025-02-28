package main

import (
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
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	logger, err := common.NewLogger("info")
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg := config.NewConfig()

	brokers := common.NewSessionManager()

	httpSrv := http.NewServer(logger, brokers, cfg.HttpListenAddr)

	bytes, err := os.ReadFile(cfg.SshPrivateKeyPath)
	if err != nil {
		logger.Error("Failed to read SSH private key",
			zap.Error(err),
		)
	}

	sshServ, err := ssh.NewServer(logger, brokers, cfg.SshListenAddr, cfg.HttpBaseURL, bytes)
	if err != nil {
		logger.Error("Failed to create SSH server",
			zap.Error(err),
		)
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			logger.Error("Failed to run HTTP server",
				zap.Error(err),
			)
		}
	}()

	go func() {
		if err := sshServ.ListenAndServe(); err != nil {
			logger.Error("Failed to run SSH server",
				zap.Error(err),
			)
		}
	}()

	<-sigChannel
}
