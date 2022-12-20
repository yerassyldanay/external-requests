package msgbroker

import "context"

// MBPublisher abstructs method(s) for publishing messages to MB
type MBPublisher interface {
	// expects message in terms of bytes
	Publish(ctx context.Context, msg []byte) error
}

// MBConsumer abstructs method(s) for consuming messages from MB
type MBConsumer interface {
	// returns the channel of messages
	Consume(ctx context.Context) (chan []byte, error)
}

// MessageBroker abtucts methods for publishing to and consuming from
// message brokers
type MessageBroker interface {
	MBPublisher
	MBConsumer
}
