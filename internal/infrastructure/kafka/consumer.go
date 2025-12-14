package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

// Consumer wraps Kafka consumer group functionality
type Consumer struct {
	consumer sarama.ConsumerGroup
}

// NewConsumer creates a new Kafka consumer group
func NewConsumer(groupID string, client sarama.Client) (*Consumer, error) {
	consumer, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, fmt.Errorf("can't create Kafka consumer: %w", err)
	}

	return &Consumer{consumer: consumer}, nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// GetConsumer returns the underlying sarama.ConsumerGroup
func (c *Consumer) GetConsumer() sarama.ConsumerGroup {
	return c.consumer
}


