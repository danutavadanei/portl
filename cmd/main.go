package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/common"
	"github.com/danutavadanei/portl/http"
	sftpext "github.com/danutavadanei/portl/sftp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	brokers := common.NewSessionManager()

	httpSrv := http.NewServer(brokers, "127.0.0.1:8090")

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to run HTTP server: %v", err)
		}
	}()

	go func() {
		if err := runSshServer(brokers); err != nil {
			log.Fatalf("Failed to run SSH server: %v", err)
		}
	}()

	<-sigChannel
}

func runSshServer(brokers *common.SessionManager) error {
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// nothing for now
			return nil, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// nothing for now
			return nil, nil
		},
		BannerCallback: func(conn ssh.ConnMetadata) string {
			sessionID := hashSessionID(conn)

			brokers.Store(sessionID, broker.NewInMemoryBroker())

			return fmt.Sprintf("http://127.0.0.1:8090/%s\n", sessionID)
		},
	}

	bytes, err := os.ReadFile("./keys/ssh.pem")
	if err != nil {
		return fmt.Errorf("failed to read private key: %v", err)
	}

	key, err := ssh.ParsePrivateKey(bytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %v", err)
	}

	cfg.AddHostKey(key)

	listener, err := net.Listen("tcp", "127.0.0.1:2222")
	if err != nil {
		return fmt.Errorf("failed to listen on 127.0.0.1:2222: %v", err)
	}

	log.Print("SSH Server listening on 127.0.0.1:2222")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}
		go handleIncomingSshConnection(conn, cfg, brokers)
	}
}

func handleIncomingSshConnection(conn net.Conn, cfg *ssh.ServerConfig, brokers *common.SessionManager) {
	shConn, chans, reqs, err := ssh.NewServerConn(conn, cfg)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
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
			log.Printf("Could not accept channel %v", err)
			return
		}

		sessionID := hashSessionID(shConn)

		b, ok := brokers.Load(sessionID)
		if !ok {
			log.Printf("Session ID not found in pipes")
			return
		}

		// block here as we don't want to handle multiple sessions concurrently
		handleSession(channel, requests, b)

		// we are done with this broker, it should never be used again
		brokers.Delete(sessionID)
	}
}

func handleSession(channel ssh.Channel, in <-chan *ssh.Request, s broker.Broker) {
	defer channel.Close()
	defer s.Close()

	for req := range in {
		switch req.Type {
		case "subsystem":
			subsystem := parseSubsystem(req.Payload)
			if subsystem == "sftp" {
				log.Println("Starting SFTP subsystem")

				handler := sftpext.NewHandler(s)

				handlers := sftp.Handlers{
					FileGet:  handler,
					FilePut:  handler,
					FileCmd:  handler,
					FileList: handler,
				}

				sftpServer := sftp.NewRequestServer(channel, handlers)

				if err := sftpServer.Serve(); err == io.EOF {
					log.Println("SFTP client exited session.")
				} else if err != nil {
					log.Printf("SFTP server completed with error: %v", err)
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
