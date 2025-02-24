package ssh

import (
	"fmt"
	"github.com/danutavadanei/portl/broker"
	"github.com/danutavadanei/portl/common"
	"golang.org/x/crypto/ssh"
)

func passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	return nil, nil
}
func publicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	return nil, nil
}

func bannerCallback(sm *common.SessionManager, httpListenAddr string) func(conn ssh.ConnMetadata) string {
	return func(conn ssh.ConnMetadata) string {
		sessionID := hashSessionID(conn)

		sm.Store(sessionID, broker.NewInMemoryBroker())

		return fmt.Sprintf("http://%s/%s\n", httpListenAddr, sessionID)
	}
}
