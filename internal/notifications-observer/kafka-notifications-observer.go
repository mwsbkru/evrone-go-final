package notifications_observer

import (
	"context"
	"encoding/json"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity"
	"evrone_course_final/internal/usecase"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

type KafkaNotificationsObserver struct {
	sarama.ConsumerGroupHandler
	topicName  string
	consumer   sarama.ConsumerGroup
	cfg        *config.Config
	subscriber usecase.NotificationsSubscriber
	terminator usecase.Terminator
}

func NewKafkaNotificationsObserver(topicName string, cfg *config.Config, consumer sarama.ConsumerGroup) *KafkaNotificationsObserver {
	return &KafkaNotificationsObserver{topicName: topicName, cfg: cfg, consumer: consumer}
}

func (k *KafkaNotificationsObserver) Subscribe(subscriber usecase.NotificationsSubscriber, terminator usecase.Terminator) {
	k.subscriber = subscriber
	k.terminator = terminator
}

func (k *KafkaNotificationsObserver) StartListening(ctx context.Context) {
	for {
		// Запуск цикла обработки сообщений
		err := k.consumer.Consume(ctx, []string{k.topicName}, k)
		if err != nil {
			slog.Error("Error consuming messages", slog.String("error", err.Error()))
		}

		select {
		case <-ctx.Done():
			slog.Info("Terminating Kafka observer")
			k.terminator()
			return
		case <-time.After(time.Duration(k.cfg.KafkaTimeoutSeconds) * time.Second):
			slog.Info(fmt.Sprintf("Listening topic %s timeout %d", k.topicName, k.cfg.KafkaTimeoutSeconds))
		}
	}
}

func (k *KafkaNotificationsObserver) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Обработка сообщений
	for message := range claim.Messages() {
		slog.Info("Kafka observer - message received", slog.String("message.key", string(message.Key)), slog.String("message.value", string(message.Value)))

		go func() {
			session.MarkMessage(message, "")
			notification, err := messageToNotification(message)
			if err != nil {
				slog.Error("KafkaNotificationsProcessor: error unmarshall json", slog.String("error", err.Error()))
				return
			}

			k.subscriber(notification)
		}()
	}

	return nil
}

func messageToNotification(message *sarama.ConsumerMessage) (*entity.Notification, error) {
	var notification entity.Notification

	err := json.Unmarshal([]byte(message.Value), &notification)
	if err != nil {
		return nil, fmt.Errorf("KafkaNotificationsProcessor error unmarshall json: %w", err)
	}

	return &notification, nil
}

func (k *KafkaNotificationsObserver) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (k *KafkaNotificationsObserver) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
