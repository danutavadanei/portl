package ssh

import (
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/danutavadanei/portl/common"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	logger     *zap.Logger
	sm         *common.SessionManager
	listenAddr string
	httpURL    string
	privateKey ssh.Signer
}

func NewServer(logger *zap.Logger, sm *common.SessionManager, listenAddr, httpURL string, privateKeyBytes []byte) (*Server, error) {
	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:     logger,
		sm:         sm,
		listenAddr: listenAddr,
		httpURL:    httpURL,
		privateKey: key,
	}, nil
}

func (s *Server) ListenAndServe() error {
	cfg := &ssh.ServerConfig{
		PasswordCallback:  passwordCallback,
		PublicKeyCallback: publicKeyCallback,
		BannerCallback:    bannerCallback(s.sm, s.httpURL),
	}

	cfg.AddHostKey(s.privateKey)

	listener, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.listenAddr, err)
	}

	s.logger.Info("SSH Server listening",
		zap.String("address", s.listenAddr),
	)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Error("Failed to accept incoming connection",
				zap.Error(err),
			)
			continue
		}
		go s.handleIncomingSshConnection(conn, cfg)
	}
}

func parseSubsystem(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	length := (uint32(payload[0]) << 24) |
		(uint32(payload[1]) << 16) |
		(uint32(payload[2]) << 8) |
		(uint32(payload[3]))
	if int(length) > len(payload)-4 {
		return ""
	}
	return string(payload[4 : 4+length])
}

func hashSessionID(conn ssh.ConnMetadata) string {
	hash := sha256.New()
	hash.Write(conn.SessionID())
	return fmt.Sprintf("%x", hash.Sum(nil))
}
