package broker

import (
	"errors"
	"sync"
)

type InMemoryBroker struct {
	mu             sync.Mutex
	queue          []Message
	consumerChan   chan Message
	consumerActive bool
}

func NewInMemoryBroker() *InMemoryBroker {
	return &InMemoryBroker{
		queue:          make([]Message, 0),
		consumerChan:   nil,
		consumerActive: false,
	}
}

func (b *InMemoryBroker) Publish(msg Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerActive && b.consumerChan != nil {
		b.consumerChan <- msg
	} else {
		b.queue = append(b.queue, msg)
	}

	return nil
}

func (b *InMemoryBroker) Subscribe() (<-chan Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerActive {
		return nil, errors.New("broker: consumer already subscribed")
	}

	b.consumerChan = make(chan Message, len(b.queue)+1)
	b.consumerActive = true

	for _, msg := range b.queue {
		b.consumerChan <- msg
	}

	b.queue = make([]Message, 0)

	return b.consumerChan, nil
}

func (b *InMemoryBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerChan != nil {
		close(b.consumerChan)
		b.consumerChan = nil
	}

	return nil
}
