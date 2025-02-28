package ssh

import (
	"fmt"
	"io"
	"net"

	"github.com/danutavadanei/portl/broker"
	sftpext "github.com/danutavadanei/portl/sftp"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func (s *Server) handleIncomingSshConnection(conn net.Conn, cfg *ssh.ServerConfig) {
	shConn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
	if err != nil {
		s.logger.Error("Failed to handshake",
			zap.Error(err),
		)
		return
	}
	defer shConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			s.logger.Error("Could not accept channel",
				zap.Error(err),
			)
			return
		}

		sessionID := hashSessionID(shConn)

		b, ok := s.sm.Load(sessionID)
		if !ok {
			s.logger.Error("Session ID not found in pipes")
			return
		}

		// block here as we don't want to handle multiple sessions concurrently
		s.handleSshSession(channel, requests, b)

		// we are done with this broker, it should never be used again
		s.sm.Delete(sessionID)
	}
}

func (s *Server) handleSshSession(channel ssh.Channel, in <-chan *ssh.Request, b broker.Broker) {
	defer channel.Close()
	defer b.Close()

	for req := range in {
		switch req.Type {
		case "subsystem":
			subsystem := parseSubsystem(req.Payload)
			if subsystem == "sftp" {
				s.logger.Info("Starting SFTP subsystem")

				handler := sftpext.NewHandler(s.logger, b)

				handlers := sftp.Handlers{
					FileGet:  handler,
					FilePut:  handler,
					FileCmd:  handler,
					FileList: handler,
				}

				sftpServer := sftp.NewRequestServer(channel, handlers)

				if err := sftpServer.Serve(); err == io.EOF {
					s.logger.Info("SFTP client exited session.")
				} else if err != nil {
					s.logger.Error("SFTP server completed with error",
						zap.Error(err),
					)
				}

				return
			}
			// If not "sftp", reject
			req.Reply(false, nil)

		default:
			// Reject all other request types
			req.Reply(false, nil)
		}
	}
}
