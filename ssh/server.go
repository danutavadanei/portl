package ssh

import (
	"crypto/sha256"
	"fmt"
	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/config"
	"net"
	"os"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	logger     *zap.Logger
	store      *broker.Store
	listenAddr string
	httpURL    string
	privateKey ssh.Signer
}

func NewServer(logger *zap.Logger, store *broker.Store, cfg *config.Config) (*Server, error) {
	privateKeyBytes, err := os.ReadFile(cfg.SshPrivateKeyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:     logger,
		store:      store,
		listenAddr: cfg.SshListenAddr,
		httpURL:    cfg.HttpBaseURL,
		privateKey: key,
	}, nil
}

func (s *Server) Serve() error {
	cfg := &ssh.ServerConfig{
		PasswordCallback:  passwordCallback,
		PublicKeyCallback: publicKeyCallback,
		BannerCallback:    bannerCallback(s.logger, s.store, s.httpURL),
	}

	cfg.AddHostKey(s.privateKey)

	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.listenAddr, err)
	}

	s.logger.Info("SSH server listening", zap.String("address", s.listenAddr))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("failed to accept incoming connection", zap.Error(err))
			continue
		}
		go s.handleIncomingSshConnection(conn, cfg)
	}
}

func hashSessionID(conn ssh.ConnMetadata) string {
	hash := sha256.New()
	hash.Write(conn.SessionID())
	return fmt.Sprintf("%x", hash.Sum(nil))
}
