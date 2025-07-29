package dead_notifications_processor

import (
	"encoding/json"
	"errors"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity"
	"fmt"
	"github.com/IBM/sarama"
	"log/slog"
)

type KafkaDeadNotificationsProcessor struct {
	producer sarama.SyncProducer
	cfg      *config.Config
}

func NewKafkaDeadNotificationsProcessor(producer sarama.SyncProducer, cfg *config.Config) *KafkaDeadNotificationsProcessor {
	return &KafkaDeadNotificationsProcessor{producer: producer, cfg: cfg}
}

// TODO добавить закрытые клиентов кафки на запись и чтение
func (k KafkaDeadNotificationsProcessor) Process(notification *entity.Notification, err error) error {
	// Проверяем, что уведомление не nil
	if notification == nil {
		return reportAndWrapErrorDeadKafka(errors.New("notification cannot be nil"))
	}

	// Создаем структуру для отправки в Kafka
	payload := struct {
		Notification *entity.Notification `json:"notification"`
		Error        string               `json:"error"`
	}{
		Notification: notification,
		Error:        err.Error(),
	}

	// Преобразуем структуру в JSON
	payloadJSON, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return reportAndWrapErrorDeadKafka(err)
	}

	// Создаем сообщение для Kafka
	msg := &sarama.ProducerMessage{
		Topic: k.cfg.KafkaTopicDeadNotifications, // имя топика Kafka
		Value: sarama.StringEncoder(payloadJSON),
	}

	// Отправляем сообщение в Kafka
	_, _, err = k.producer.SendMessage(msg)
	if err != nil {
		return reportAndWrapErrorDeadKafka(err)
	}

	return nil
}

func reportAndWrapErrorDeadKafka(err error) error {
	slog.Error("DeadKafkaNotificationsProcessor error", slog.String("error", err.Error()))
	return fmt.Errorf("DeadKafkaNotificationsProcessor error: %w", err)
}
