package ssh

import (
	"fmt"
	"github.com/danutavadanei/portl/broker"
	"go.uber.org/zap"
	"time"

	"github.com/danutavadanei/portl/common"
	"golang.org/x/crypto/ssh"
)

func passwordCallback(_ ssh.ConnMetadata, _ []byte) (*ssh.Permissions, error) {
	return nil, nil
}
func publicKeyCallback(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) {
	return nil, nil
}

func bannerCallback(l *zap.Logger, sm *common.SessionManager, httpURL string) func(conn ssh.ConnMetadata) string {
	return func(conn ssh.ConnMetadata) string {
		sessionID := hashSessionID(conn)
		b := broker.NewInMemoryBroker()
		sm.Store(sessionID, b)

		err := conn.(ssh.ServerPreAuthConn).SendAuthBanner(fmt.Sprintf("Share the download link:\n%s/%s\nThis link will be active for 5 minutes.\n", httpURL, sessionID))
		if err != nil {
			l.Error("failed to send download auth banner", zap.Error(err))
			return ""
		}

		subscribedChan := b.WaitForSubscription()
		tickerChan := time.After(5 * time.Minute)

		for {
			select {
			case <-tickerChan:
				if err = conn.(ssh.ServerPreAuthConn).SendAuthBanner("Your session has expired.\n"); err != nil {
					l.Error("failed to send expired auth banner", zap.Error(err))
				}

				if err = conn.(ssh.Conn).Close(); err != nil {
					l.Error("failed to close connection", zap.Error(err))
				}
				return ""
			case <-subscribedChan:
				return "Connected to the peer, starting transfer...\n"
			}
		}
	}
}
