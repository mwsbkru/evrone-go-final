package async_notifications

import (
	"context"
	"evrone_course_final/config"
	dead_notifications_processor "evrone_course_final/internal/dead-notifications-processor"
	notifications_observer "evrone_course_final/internal/notifications-observer"
	notifications_processor "evrone_course_final/internal/notifications-processor"
	"evrone_course_final/internal/tools"
	"evrone_course_final/internal/usecase"
	"github.com/IBM/sarama"
	"log/slog"
)

func Run(ctx context.Context, cfg *config.Config) {
	kafkaClient, err := tools.PrepareKafkaClient(cfg)
	if err != nil {
		slog.Error("Can`t init Kafka client", slog.String("error", err.Error()))
		return
	}

	consumerEmail, err := sarama.NewConsumerGroupFromClient(cfg.KafkaConsumerGroupID, kafkaClient)
	if err != nil {
		slog.Error("Can`t init Kafka consumerEmail", slog.String("error", err.Error()))
		return
	}

	consumerPush, err := sarama.NewConsumerGroupFromClient(cfg.KafkaConsumerGroupID, kafkaClient)
	if err != nil {
		slog.Error("Can`t init Kafka consumerPush", slog.String("error", err.Error()))
		return
	}

	topicEmailNotifications := cfg.KafkaTopicEmailNotifications
	kafkaObserverEmail := notifications_observer.NewKafkaNotificationsObserver(topicEmailNotifications, cfg, consumerEmail)

	// TODO: Добавить в dead notifications processor отправкку в кафку
	processorEmail := notifications_processor.NewEmailNotificationsProcessor(cfg)
	deadProcessorEmail := dead_notifications_processor.ConsoleDeadNotificationsProcessor{Name: "email"}
	notificationsChannelEmail := usecase.NewNotificationChannelUseCase(cfg, "Email processor", kafkaObserverEmail, processorEmail, &deadProcessorEmail)

	topicPushNotifications := cfg.KafkaTopicPushNotifications
	kafkaObserverPush := notifications_observer.NewKafkaNotificationsObserver(topicPushNotifications, cfg, consumerPush)

	consoleProcessorPush := notifications_processor.ConsoleNotificationsProcessor{Name: "push"}
	deadProcessorPush := dead_notifications_processor.ConsoleDeadNotificationsProcessor{Name: "push"}
	notificationsChannelPush := usecase.NewNotificationChannelUseCase(cfg, "Push processor", kafkaObserverPush, &consoleProcessorPush, &deadProcessorPush)

	notificationsChannels := []*usecase.NotificationsChannelUseCase{notificationsChannelEmail, notificationsChannelPush}
	notificationsUseCase := usecase.NewNotificationsUseCase(notificationsChannels)
	notificationsUseCase.Run(ctx)
}
