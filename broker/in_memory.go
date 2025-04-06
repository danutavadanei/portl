package broker

import (
	"errors"
	"sync"
)

type InMemory struct {
	mu             sync.Mutex
	queue          []Message
	consumerChan   chan Message
	consumerActive bool
	waitChan       chan struct{}
}

func NewInMemory() *InMemory {
	return &InMemory{
		queue:          make([]Message, 0),
		consumerChan:   nil,
		consumerActive: false,
		waitChan:       make(chan struct{}),
	}
}

func (b *InMemory) Publish(msg Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerActive && b.consumerChan != nil {
		b.consumerChan <- msg
	} else {
		b.queue = append(b.queue, msg)
	}

	return nil
}

func (b *InMemory) Subscribe() (<-chan Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerActive {
		return nil, errors.New("broker: consumer already subscribed")
	}

	b.consumerChan = make(chan Message, len(b.queue)+1)
	b.consumerActive = true
	b.waitChan <- struct{}{}

	for _, msg := range b.queue {
		b.consumerChan <- msg
	}

	b.queue = make([]Message, 0)

	return b.consumerChan, nil
}

func (b *InMemory) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.consumerChan != nil {
		close(b.consumerChan)
		b.consumerChan = nil
	}

	return nil
}

func (b *InMemory) WaitForSubscription() <-chan struct{} {
	return b.waitChan
}
