package async_notifications

import (
	"context"
	"evrone_course_final/cmd/config"
	dead_notifications_processor "evrone_course_final/internal/dead-notifications-processor"
	notifications_observer "evrone_course_final/internal/notifications-observer"
	notifications_processor "evrone_course_final/internal/notifications-processor"
	"evrone_course_final/internal/usecase"
	"github.com/IBM/sarama"
	"log/slog"
	"time"
)

func Run(ctx context.Context, cfg *config.Config) {
	kafkaClient, err := prepareKafkaClient(cfg)
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

	consoleProcessorEmail := notifications_processor.ConsoleNotificationsProcessor{Name: "email"}
	deadProcessorEmail := dead_notifications_processor.ConsoleDeadNotificationsProcessor{Name: "email"}
	notificationsChannelEmail := usecase.NewNotificationChannelUseCase("Email processor", kafkaObserverEmail, &consoleProcessorEmail, &deadProcessorEmail)

	topicPushNotifications := cfg.KafkaTopicPushNotifications
	kafkaObserverPush := notifications_observer.NewKafkaNotificationsObserver(topicPushNotifications, cfg, consumerPush)

	consoleProcessorPush := notifications_processor.ConsoleNotificationsProcessor{Name: "push"}
	deadProcessorPush := dead_notifications_processor.ConsoleDeadNotificationsProcessor{Name: "push"}
	notificationsChannelPush := usecase.NewNotificationChannelUseCase("Push processor", kafkaObserverPush, &consoleProcessorPush, &deadProcessorPush)

	ff(notificationsChannelPush)

	notificationsChannels := []*usecase.NotificationsChannelUseCase{notificationsChannelEmail, notificationsChannelPush}
	notificationsUseCase := usecase.NewNotificationsUseCase(notificationsChannels)
	notificationsUseCase.Run(ctx)
}

func ff(useCase *usecase.NotificationsChannelUseCase) {

}

func prepareKafkaClient(cfg *config.Config) (sarama.Client, error) {
	brokers := []string{cfg.KafkaBrokers}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Consumer.Group.Session.Timeout = time.Duration(cfg.KafkaTimeoutSeconds) * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(cfg.KafkaIntervalSeconds) * time.Second
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	return sarama.NewClient(brokers, kafkaConfig)
}
