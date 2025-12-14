package async_notifications

import (
	"context"
	"fmt"

	"github.com/mwsbkru/evrone-go-final/config"
	dead_notifications_processor "github.com/mwsbkru/evrone-go-final/internal/dead-notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/infrastructure/kafka"
	"github.com/mwsbkru/evrone-go-final/internal/infrastructure/smtp"
	notifications_observer "github.com/mwsbkru/evrone-go-final/internal/notifications-observer"
	notifications_processor "github.com/mwsbkru/evrone-go-final/internal/notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/service"
)

func Run(ctx context.Context, cfg *config.Config) error {
	kafkaClient, err := kafka.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("can't init Kafka client: %w", err)
	}
	defer kafkaClient.Close()

	producer, err := kafka.NewProducer([]string{cfg.Kafka.Brokers})
	if err != nil {
		return fmt.Errorf("can't init Kafka producer: %w", err)
	}
	defer producer.Close()

	deadProcessor := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer.GetProducer(), cfg)

	smtpClient, err := smtp.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("can't initialize SMTP client: %w", err)
	}
	defer smtpClient.Close()

	notificationsChannelEmail, consumerEmail, err := initializeEmailNotificationChannel(cfg, kafkaClient, smtpClient, deadProcessor)
	if err != nil {
		return fmt.Errorf("can't initialize email notification channel: %w", err)
	}
	defer consumerEmail.Close()

	notificationsChannelPush, consumerPush, err := initializePushNotificationChannel(cfg, kafkaClient, deadProcessor)
	if err != nil {
		return fmt.Errorf("can't initialize push notification channel: %w", err)
	}
	defer consumerPush.Close()

	notificationsChannels := []*service.NotificationsChannel{notificationsChannelEmail, notificationsChannelPush}
	notificationsService := service.NewNotificationsService(notificationsChannels)
	notificationsService.Run(ctx)

	return nil
}

// initializeEmailNotificationChannel creates and initializes an email notification channel
func initializeEmailNotificationChannel(
	cfg *config.Config,
	kafkaClient *kafka.Client,
	smtpClient *smtp.Client,
	deadProcessor service.DeadNotificationsProcessor,
) (*service.NotificationsChannel, *kafka.Consumer, error) {
	consumerEmail, err := kafka.NewConsumer(cfg.Kafka.ConsumerGroupID, kafkaClient.GetClient())
	if err != nil {
		return nil, nil, fmt.Errorf("can't init Kafka consumerEmail: %w", err)
	}

	topicEmailNotifications := cfg.Kafka.TopicEmailNotifications
	kafkaObserverEmail := notifications_observer.NewKafkaNotificationsObserver(topicEmailNotifications, cfg, consumerEmail.GetConsumer())
	processorEmail := notifications_processor.NewEmailNotificationsProcessor(cfg, smtpClient.GetClient())
	channel := service.NewNotificationChannel(cfg, "Email processor", kafkaObserverEmail, processorEmail, deadProcessor)
	return channel, consumerEmail, nil
}

// initializePushNotificationChannel creates and initializes a push notification channel
func initializePushNotificationChannel(
	cfg *config.Config,
	kafkaClient *kafka.Client,
	deadProcessor service.DeadNotificationsProcessor,
) (*service.NotificationsChannel, *kafka.Consumer, error) {
	consumerPush, err := kafka.NewConsumer(cfg.Kafka.ConsumerGroupID, kafkaClient.GetClient())
	if err != nil {
		return nil, nil, fmt.Errorf("can't init Kafka consumerPush: %w", err)
	}

	topicPushNotifications := cfg.Kafka.TopicPushNotifications
	kafkaObserverPush := notifications_observer.NewKafkaNotificationsObserver(topicPushNotifications, cfg, consumerPush.GetConsumer())
	consoleProcessorPush := notifications_processor.ConsoleNotificationsProcessor{Name: "push"}
	channel := service.NewNotificationChannel(cfg, "Push processor", kafkaObserverPush, &consoleProcessorPush, deadProcessor)
	return channel, consumerPush, nil
}
