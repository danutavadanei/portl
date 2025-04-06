package broker

import (
	"sync"
)

type Store struct {
	brokers sync.Map
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Load(sessionID string) (Broker, bool) {
	b, ok := s.brokers.Load(sessionID)
	if !ok {
		return nil, false
	}
	return b.(Broker), true
}

func (s *Store) Store(sessionID string, b Broker) {
	s.brokers.Store(sessionID, b)
}

func (s *Store) Delete(sessionID string) {
	s.brokers.Delete(sessionID)
}
