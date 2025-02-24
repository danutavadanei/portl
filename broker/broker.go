package broker

type MessageType int

const (
	Mkdir MessageType = iota
	Put
)

type Message struct {
	Type MessageType
	Path string
	Data chan []byte
}

type Broker interface {
	Publish(msg Message) error
	Subscribe() (<-chan Message, error)
	Close() error
}
