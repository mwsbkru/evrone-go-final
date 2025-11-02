package ws_notifications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mwsbkru/evrone-go-final/config"
	"github.com/mwsbkru/evrone-go-final/internal/controller/http"
	dead_notifications_processor "github.com/mwsbkru/evrone-go-final/internal/dead-notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/infrastructure/kafka"
	"github.com/mwsbkru/evrone-go-final/internal/infrastructure/redis"
	notifications_observer "github.com/mwsbkru/evrone-go-final/internal/notifications-observer"
	notifications_processor "github.com/mwsbkru/evrone-go-final/internal/notifications-processor"
	"github.com/mwsbkru/evrone-go-final/internal/service"
	"github.com/mwsbkru/evrone-go-final/internal/tools"
	ws_notifications_receivers "github.com/mwsbkru/evrone-go-final/internal/ws-notifications-receivers"
)

func Run(ctx context.Context, cfg *config.Config) error {
	kafkaClient, err := kafka.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("can't init Kafka client: %w", err)
	}
	defer kafkaClient.Close()

	redisClient := redis.NewClient(cfg)
	defer redisClient.Close()

	notificationsChannelWs, consumerWs, producer, err := initializeWSNotificationChannel(cfg, kafkaClient, redisClient)
	if err != nil {
		return fmt.Errorf("can't initialize WS notification channel: %w", err)
	}
	defer consumerWs.Close()
	defer producer.Close()

	notificationsChannels := []*service.NotificationsChannel{notificationsChannelWs}
	notificationsService := service.NewNotificationsService(notificationsChannels)
	go notificationsService.Run(ctx)

	slog.Info("Starting http server...")
	wsNotificationsReceiver := ws_notifications_receivers.NewRedisWsNotificationsReceiver(redisClient.GetClient(), cfg)
	wsNotificationsService := service.NewWsNotificationsService(wsNotificationsReceiver)
	wsNotificationsService.Run(ctx)

	server := http.NewServer(cfg, wsNotificationsService)
	http.Serve(ctx, server, cfg)

	return nil
}

// initializeWSNotificationChannel creates and initializes a WebSocket notification channel
func initializeWSNotificationChannel(
	cfg *config.Config,
	kafkaClient *kafka.Client,
	redisClient *redis.Client,
) (*service.NotificationsChannel, *kafka.Consumer, *kafka.Producer, error) {
	consumerWs, err := kafka.NewConsumer(cfg.Kafka.ConsumerGroupID, kafkaClient.GetClient())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't init Kafka WebSocket consumer: %w", err)
	}

	err = tools.EnsureTopicExists(cfg.Kafka.TopicDeadNotifications, kafkaClient.GetClient())
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't prepare topic for dead notifications: %w", err)
	}

	producer, err := kafka.NewProducer([]string{cfg.Kafka.Brokers})
	if err != nil {
		consumerWs.Close()
		return nil, nil, nil, fmt.Errorf("can't init Kafka producer: %w", err)
	}

	topicWsNotifications := cfg.Kafka.TopicWSNotifications
	kafkaObserverWs := notifications_observer.NewKafkaNotificationsObserver(topicWsNotifications, cfg, consumerWs.GetConsumer())
	processorWs := notifications_processor.NewRedisWSNotificationsProcessor(redisClient.GetClient())
	deadProcessorWs := dead_notifications_processor.NewKafkaDeadNotificationsProcessor(producer.GetProducer(), cfg)
	channel := service.NewNotificationChannelUseCase(cfg, "WS processor", kafkaObserverWs, processorWs, deadProcessorWs)
	return channel, consumerWs, producer, nil
}
