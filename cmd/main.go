package main

import (
	"github.com/danutavadanei/portl/common"
	"github.com/danutavadanei/portl/config"
	"github.com/danutavadanei/portl/http"
	"github.com/danutavadanei/portl/ssh"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	cfg := config.NewConfig()

	brokers := common.NewSessionManager()

	httpSrv := http.NewServer(brokers, cfg.HttpListenAddr)

	bytes, err := os.ReadFile(cfg.SshPrivateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read SSH private key: %v", err)
	}

	sshServ, err := ssh.NewServer(brokers, cfg.SshListenAddr, cfg.HttpBaseURL, bytes)
	if err != nil {
		log.Fatalf("Failed to create SSH server: %v", err)
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to run HTTP server: %v", err)
		}
	}()

	go func() {
		if err := sshServ.ListenAndServe(); err != nil {
			log.Fatalf("Failed to run SSH server: %v", err)
		}
	}()

	<-sigChannel
}
