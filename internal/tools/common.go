package tools

import (
	"evrone_course_final/config"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

func PrepareKafkaClient(cfg *config.Config) (sarama.Client, error) {
	brokers := []string{cfg.Kafka.Brokers}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Group.Session.Timeout = time.Duration(cfg.Kafka.TimeoutSeconds) * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(cfg.Kafka.IntervalSeconds) * time.Second
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return sarama.NewClient(brokers, kafkaConfig)
}

func topicExists(admin sarama.ClusterAdmin, topic string) (bool, error) {
	topics, err := admin.ListTopics()
	if err != nil {
		return false, fmt.Errorf("can`t fetch list of topics: %w", err)
	}

	_, exists := topics[topic]
	return exists, nil
}

func createTopic(admin sarama.ClusterAdmin, topic string, partitions int32, replicationFactor int16) error {
	topicConfig := sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	err := admin.CreateTopic(topic, &topicConfig, false)
	if err != nil {
		return fmt.Errorf("cant create topic: %w", err)
	}

	return nil
}

func EnsureTopicExists(topic string, client sarama.Client) error {
	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return fmt.Errorf("can`t ensure topic exists: %w", err)
	}

	exists, err := topicExists(admin, topic)
	if err != nil {
		return fmt.Errorf("can`t check topic exists: %w", err)
	}

	if !exists {
		slog.Info("topic not exists, creating")
		// тут пока не стал упарываться в проброс этих параметров в конфиг
		err = createTopic(admin, topic, 3, 1)
		if err != nil {
			return fmt.Errorf("can`t create new topic: %w", err)
		}

		// Ждем некоторое время, чтобы топик был создан
		time.Sleep(2 * time.Second)
	}

	return nil
}
