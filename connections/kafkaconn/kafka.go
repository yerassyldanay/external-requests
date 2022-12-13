package kafkaconn

import (
	"time"

	"github.com/segmentio/kafka-go"
)

func NewMBWriter(addresses []string, topic string) *kafka.Writer {
	writer := kafka.NewWriter(
		kafka.WriterConfig{
			Brokers:  addresses,
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	)
	writer.AllowAutoTopicCreation = true

	return writer
}

func NewMBReader(addresses []string, topic string, partition int) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          addresses,
		GroupID:          "consumer-group-id",
		Topic:            topic,
		ReadBatchTimeout: time.Millisecond * 10,
		CommitInterval:   time.Millisecond * 1,
		Partition:        partition,
	})

	return reader
}
