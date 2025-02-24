package common

import (
	"sync"

	"github.com/danutavadanei/portl/broker"
)

type SessionManager struct {
	brokers sync.Map
}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (sm *SessionManager) Load(sessionID string) (broker.Broker, bool) {
	b, ok := sm.brokers.Load(sessionID)
	if !ok {
		return nil, false
	}
	return b.(broker.Broker), true
}

func (sm *SessionManager) Store(sessionID string, b broker.Broker) {
	sm.brokers.Store(sessionID, b)
}

func (sm *SessionManager) Delete(sessionID string) {
	sm.brokers.Delete(sessionID)
}
