package async_notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/mwsbkru/evrone-go-final/config"
	dead_notifications_processor "github.com/mwsbkru/evrone-go-final/internal/dead-notifications-processor"
	notifications_observer "github.com/mwsbkru/evrone-go-final/internal/notifications-observer"
	notifications_processor "github.com/mwsbkru/evrone-go-final/internal/notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/mwsbkru/evrone-go-final/internal/tools"

	"github.com/IBM/sarama"
	mail "github.com/xhit/go-simple-mail/v2"
)

func Run(ctx context.Context, cfg *config.Config) error {
	kafkaClient, err := tools.PrepareKafkaClient(cfg)
	if err != nil {
		return fmt.Errorf("can't init Kafka client: %w", err)
	}
	defer kafkaClient.Close()

	consumerEmail, err := sarama.NewConsumerGroupFromClient(cfg.Kafka.ConsumerGroupID, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't init Kafka consumerEmail: %w", err)
	}
	defer consumerEmail.Close()

	consumerPush, err := sarama.NewConsumerGroupFromClient(cfg.Kafka.ConsumerGroupID, kafkaClient)
	if err != nil {
		return fmt.Errorf("can't init Kafka consumerPush: %w", err)
	}
	defer consumerPush.Close()

	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer([]string{cfg.Kafka.Brokers}, producerConfig)
	if err != nil {
		return fmt.Errorf("can't init Kafka producer: %w", err)
	}
	defer producer.Close()

	deadProcessor := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer, cfg)

	smtpClient, err := initializeSMTPClient(cfg)
	if err != nil {
		return fmt.Errorf("can't initialize SMTP client: %w", err)
	}
	defer smtpClient.Close()

	topicEmailNotifications := cfg.Kafka.TopicEmailNotifications
	kafkaObserverEmail := notifications_observer.NewKafkaNotificationsObserver(topicEmailNotifications, cfg, consumerEmail)
	processorEmail := notifications_processor.NewEmailNotificationsProcessor(cfg, smtpClient)
	notificationsChannelEmail := service.NewNotificationChannelUseCase(cfg, "Email processor", kafkaObserverEmail, processorEmail, deadProcessor)

	topicPushNotifications := cfg.Kafka.TopicPushNotifications
	kafkaObserverPush := notifications_observer.NewKafkaNotificationsObserver(topicPushNotifications, cfg, consumerPush)
	consoleProcessorPush := notifications_processor.ConsoleNotificationsProcessor{Name: "push"}
	notificationsChannelPush := service.NewNotificationChannelUseCase(cfg, "Push processor", kafkaObserverPush, &consoleProcessorPush, deadProcessor)

	notificationsChannels := []*service.NotificationsChannel{notificationsChannelEmail, notificationsChannelPush}
	notificationsService := service.NewNotificationsService(notificationsChannels)
	notificationsService.Run(ctx)

	return nil
}

// initializeSMTPClient creates and configures an SMTP client based on the provided configuration
func initializeSMTPClient(cfg *config.Config) (*mail.SMTPClient, error) {
	server := mail.NewSMTPClient()
	server.Host = cfg.Email.SmtpServerHost
	server.Port = cfg.Email.SmtpServerPort
	server.Username = cfg.Email.SmtpUsername
	server.Password = cfg.Email.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS

	server.KeepAlive = true

	server.ConnectTimeout = time.Duration(cfg.Email.SmtpTimeoutSeconds) * time.Second
	server.SendTimeout = time.Duration(cfg.Email.SmtpTimeoutSeconds) * time.Second

	return server.Connect()
}
