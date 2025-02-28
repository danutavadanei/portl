package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/danutavadanei/portl/common"
	"github.com/danutavadanei/portl/config"
	"github.com/danutavadanei/portl/http"
	"github.com/danutavadanei/portl/ssh"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logLevel := zapcore.InfoLevel
	if *debug {
		logLevel = zapcore.DebugLevel
	}

	logger, err := common.NewLogger(logLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

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
