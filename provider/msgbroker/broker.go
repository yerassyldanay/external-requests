package msgbroker

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MBConnection wraps the connection to message broker
type MBConnection struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
	logger *zap.Logger
}

var _ MessageBroker = (*MBConnection)(nil)

func NewMBConnection(writer *kafka.Writer, reader *kafka.Reader, logger *zap.Logger) (*MBConnection, error) {
	return &MBConnection{
		Reader: reader,
		Writer: writer,
		logger: logger,
	}, nil
}

// Publish accepts message as a byte and sends it to message broker under the hood
func (mb *MBConnection) Publish(ctx context.Context, msg []byte) error {
	mb.logger.Debug("send message to message broker", zap.ByteString("msg", msg))
	err := mb.Writer.WriteMessages(ctx, kafka.Message{Value: msg})
	if err != nil {
		return fmt.Errorf("failed to write msg to MB. err: %v", err)
	}
	return nil
}

// Consume runs the listener under the hood and returns channel.
// Every message consumed from the message broker will be sent to this channel
func (mb *MBConnection) Consume(ctx context.Context) (chan []byte, error) {
	var ch = make(chan []byte)
	go func() {

		defer close(ch)
		defer mb.logger.Info("stop consumer from running...")

	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			default:
			}

			msg, err := mb.Reader.ReadMessage(ctx)
			if err != nil {
				mb.logger.Error("failed to read msg", zap.Error(err))
				continue
			}

			mb.logger.Debug("received a new message", zap.ByteString("msg", msg.Value))

			ch <- msg.Value
		}
	}()

	return ch, nil
}

func (mb *MBConnection) Close() {
	if mb == nil {
		return
	}

	if mb.Reader != nil {
		mb.Reader.Close()
	}

	if mb.Writer != nil {
		mb.Writer.Close()
	}
}
