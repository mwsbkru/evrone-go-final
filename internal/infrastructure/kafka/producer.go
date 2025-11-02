package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

// Producer wraps Kafka producer functionality
type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new Kafka sync producer
func NewProducer(brokers []string) (*Producer, error) {
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(brokers, producerConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create Kafka producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// GetProducer returns the underlying sarama.SyncProducer
func (p *Producer) GetProducer() sarama.SyncProducer {
	return p.producer
}
