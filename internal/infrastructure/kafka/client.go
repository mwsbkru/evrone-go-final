package kafka

import (
	"fmt"
	"time"

	"github.com/mwsbkru/evrone-go-final/config"

	"github.com/IBM/sarama"
)

// Client wraps Kafka client functionality
type Client struct {
	client sarama.Client
}

// NewClient creates a new Kafka client
func NewClient(cfg *config.Config) (*Client, error) {
	brokers := []string{cfg.Kafka.Brokers}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Group.Session.Timeout = time.Duration(cfg.Kafka.TimeoutSeconds) * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(cfg.Kafka.IntervalSeconds) * time.Second
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewClient(brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create Kafka client: %w", err)
	}

	return &Client{client: client}, nil
}

// Close closes the Kafka client
func (c *Client) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying sarama.Client
func (c *Client) GetClient() sarama.Client {
	return c.client
}


