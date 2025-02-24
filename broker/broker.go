package broker

import "io"

type MessageType int

const (
	Mkdir MessageType = iota
	Put
)

type Message struct {
	Type MessageType
	Path string
	Data io.ReadCloser
}

type Broker interface {
	Publish(msg Message) error
	Subscribe() (<-chan Message, error)
	Close() error
}
