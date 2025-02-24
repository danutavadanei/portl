package main

import (
	"github.com/danutavadanei/portl/common"
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

	brokers := common.NewSessionManager()

	httpSrv := http.NewServer(brokers, "127.0.0.1:8090")

	bytes, err := os.ReadFile("./keys/ssh.pem")
	if err != nil {
		log.Fatalf("Failed to read SSH private key: %v", err)
	}

	sshServ, err := ssh.NewServer(brokers, "127.0.0.1:2222", "127.0.0.1:8090", bytes)
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
